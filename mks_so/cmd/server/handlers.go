package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"mks_sql/pkg/config"
	"mks_sql/pkg/processor"
	"mks_sql/pkg/query"

	"github.com/jackc/pgx/v5"
)

var activeDB *pgx.Conn

type ProcessRequest struct {
	SqlText string `json:"sql"`
	Input   string `json:"input"`
	Minify  bool   `json:"minify"`
}

type ProcessResponse struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, ProcessResponse{Error: "Invalid JSON input"}, http.StatusBadRequest)
		return
	}

	// Basic validation of Input JSON to ensure it's valid JSON string
	// The processor takes a JSON string, so we just pass it valid or not?
	// The request 'Input' is a string that SHOULD be a JSON object literal.
	// But let's let the processor handle it or just pass it through.
	// Actually, req.Input is a string.

	result := processor.ProcessSql(req.SqlText, req.Input, req.Minify)
	jsonResponse(w, ProcessResponse{Result: result}, http.StatusOK)
}

func handleTests(w http.ResponseWriter, r *http.Request) {
	tests := config.LoadTests(nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tests)
}

func handleRules(w http.ResponseWriter, r *http.Request) {
	patterns, err := config.LoadConfigs(nil)
	if err != nil {
		http.Error(w, "Failed to load rules", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patterns)
}

func handleDoc(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig(nil)
	docPath := appCfg.ReferenceDocPath
	if docPath == "" {
		docPath = "doc/reference_guide.md"
	}
	content, err := os.ReadFile(docPath)
	if err != nil {
		// Try fallback locations
		fallbacks := []string{"../" + docPath, "../../../" + docPath, "doc/reference_guide.md"}
		for _, f := range fallbacks {
			content, err = os.ReadFile(f)
			if err == nil {
				break
			}
		}
		if err != nil {
			http.Error(w, "Documentation not found", http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.Write(content)
}

func handleParserRules(w http.ResponseWriter, r *http.Request) {
	appCfg := config.LoadAppConfig(nil)
	docPath := appCfg.ParserRulesDocPath
	if docPath == "" {
		docPath = "doc/parser_rules.md"
	}
	content, err := os.ReadFile(docPath)
	if err != nil {
		fallbacks := []string{"../" + docPath, "../../../" + docPath, "doc/parser_rules.md"}
		for _, f := range fallbacks {
			content, err = os.ReadFile(f)
			if err == nil {
				break
			}
		}
		if err != nil {
			http.Error(w, "Parser rules not found", http.StatusNotFound)
			return
		}
	}
	w.Header().Set("Content-Type", "text/markdown")
	w.Write(content)
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	version, lastBuild := config.LoadBuildInfo(nil)
	appCfg := config.LoadAppConfig(nil)
	jsonResponse(w, map[string]interface{}{
		"version":    version,
		"last_build": lastBuild,
		"sql_limits": appCfg.SqlLimits,
	}, http.StatusOK)
}

func handleDatabases(w http.ResponseWriter, r *http.Request) {
	dbs := config.LoadDatabaseConfigs(nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbs)
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg config.DatabaseConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		jsonResponse(w, map[string]string{"error": "Invalid JSON input"}, http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Disconnect existing if any
	if activeDB != nil {
		activeDB.Close(ctx)
	}

	// Build connection string with search_path and read-only mode
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s options='-c default_transaction_read_only=on'",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.Schema)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		jsonResponse(w, map[string]string{"error": fmt.Sprintf("Failed to connect: %v", err)}, http.StatusInternalServerError)
		return
	}

	activeDB = conn
	jsonResponse(w, map[string]string{"message": "Connected successfully", "name": cfg.Name}, http.StatusOK)
}

func handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if activeDB != nil {
		activeDB.Close(context.Background())
		activeDB = nil
	}
	jsonResponse(w, map[string]string{"message": "Disconnected"}, http.StatusOK)
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if activeDB == nil {
		jsonResponse(w, map[string]string{"error": "No active database connection"}, http.StatusBadRequest)
		return
	}

	var req struct {
		Sql    string          `json:"sql"`
		Params json.RawMessage `json:"params"`
		Order  []string        `json:"order"`
		Limit  int             `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, map[string]string{"error": "Invalid request"}, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var rows pgx.Rows
	var err error

	// Determine if parameterized query should be used
	hasPlaceholders := false
	for i := 1; i <= 10; i++ {
		if strings.Contains(req.Sql, fmt.Sprintf("$%d", i)) {
			hasPlaceholders = true
			break
		}
	}

	// Prepare arguments
	var args []any
	if len(req.Params) > 0 {
		if len(req.Order) > 0 {
			args, _ = query.JsonParamsToArgs(req.Params, req.Order)
		} else if hasPlaceholders {
			// Default for MKS: $1 is the whole JSON object
			args = append(args, string(req.Params))
		}
	}

	appCfg := config.LoadAppConfig(nil)

	// Substituted SQL for debugging and COPY mode
	substitutedSQL := query.SubstituteParameters(req.Sql, args...)

	// Construct "Launched SQL" and final execution SQL
	var launchSql string
	var execSql string

	if appCfg.SqlExecuteMode == "COPY" {
		// COPY mode behavior (as per workflow)
		launchSql = substitutedSQL
		if req.Limit > 0 {
			launchSql = fmt.Sprintf("%s\nLIMIT %d", launchSql, req.Limit)
		}
		execSql = launchSql
	} else {
		// EXECUTE mode formula: wrap in subquery
		wrappedSql := fmt.Sprintf("SELECT * FROM (\n%s\n) AS mks_wrapper", req.Sql)

		// Setup args for EXECUTE: $1=json, $2=limit
		paramsStr := string(req.Params)
		if paramsStr == "" || paramsStr == "null" {
			paramsStr = "{}"
		}
		args = []any{paramsStr}
		hasPlaceholders = true

		if req.Limit > 0 {
			wrappedSql = fmt.Sprintf("%s\nLIMIT $2", wrappedSql)
			args = append(args, req.Limit)
		}

		// Replace :limit and $limit placeholders in the wrapped result with $2
		wrappedSql = regexp.MustCompile(`(?i):limit|\$limit`).ReplaceAllString(wrappedSql, "$2")

		// Cast $1 to jsonb as per workflow and user's snippet
		execSql = strings.ReplaceAll(wrappedSql, "$1", "$1::jsonb")

		// Construct launchSql for display (show actual limit value for clarity)
		displaySql := strings.ReplaceAll(wrappedSql, "$1", "$1::jsonb")
		if req.Limit > 0 {
			displaySql = strings.ReplaceAll(displaySql, "$2", fmt.Sprintf("%d", req.Limit))
		}
		safeParams := strings.ReplaceAll(paramsStr, "'", "''")
		launchSql = fmt.Sprintf("EXECUTE (\n%s\n) USING jsonb '%s'", displaySql, safeParams)
	}

	if len(args) > 0 && hasPlaceholders && appCfg.SqlExecuteMode != "COPY" {
		rows, err = activeDB.Query(ctx, execSql, args...)
	} else {
		rows, err = activeDB.Query(ctx, execSql)
	}

	if err != nil {
		jsonResponse(w, map[string]interface{}{
			"error":           fmt.Sprintf("Query failed: %v", err),
			"substituted_sql": launchSql,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	cols := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		cols[i] = string(fd.Name)
	}

	var resultRows = [][]interface{}{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			jsonResponse(w, map[string]string{"error": fmt.Sprintf("Scan failed: %v", err)}, http.StatusInternalServerError)
			return
		}

		row := make([]interface{}, len(cols))
		for i := range cols {
			val := values[i]
			// pgx returns various types. For JSON output, we might need to convert some (like []byte)
			if b, ok := val.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = val
			}
		}
		resultRows = append(resultRows, row)
	}

	jsonResponse(w, map[string]interface{}{
		"columns":         cols,
		"rows":            resultRows,
		"substituted_sql": launchSql,
	}, http.StatusOK)
}

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
