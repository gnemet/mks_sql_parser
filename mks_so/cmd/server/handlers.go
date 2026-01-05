package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	jsonResponse(w, map[string]string{
		"version":    version,
		"last_build": lastBuild,
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

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s",
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
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, map[string]string{"error": "Invalid request"}, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var rows pgx.Rows
	var err error

	// If Order is missing but Params are present, generate Order from map keys
	if len(req.Params) > 0 && len(req.Order) == 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(req.Params, &m); err == nil {
			for k := range m {
				req.Order = append(req.Order, k)
			}
		}
	}

	appCfg := config.LoadAppConfig(nil)

	// Substituted SQL for debugging
	var substitutedSQL string
	if len(req.Params) > 0 && len(req.Order) > 0 {
		args, _ := query.JsonParamsToArgs(req.Params, req.Order)
		substitutedSQL = query.SubstituteParameters(req.Sql, args...)
	} else {
		substitutedSQL = strings.ReplaceAll(req.Sql, "$1", "$1::jsonb")
	}

	if len(req.Params) > 0 && len(req.Order) > 0 {
		if appCfg.SqlExecuteMode == "COPY" {
			// In COPY mode, we execute the substituted SQL directly
			rows, err = activeDB.Query(ctx, substitutedSQL)
		} else {
			// In EXECUTE mode (default), we use parameterized query
			rows, err = query.QueryJSON(ctx, activeDB, req.Sql, req.Params, req.Order)
		}
	} else {
		rows, err = activeDB.Query(ctx, substitutedSQL)
	}

	if err != nil {
		jsonResponse(w, map[string]interface{}{
			"error":           fmt.Sprintf("Query failed: %v", err),
			"substituted_sql": substitutedSQL,
		}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	cols := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		cols[i] = string(fd.Name)
	}

	var result = []map[string]interface{}{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			jsonResponse(w, map[string]string{"error": fmt.Sprintf("Scan failed: %v", err)}, http.StatusInternalServerError)
			return
		}

		row := make(map[string]interface{})
		for i, colName := range cols {
			val := values[i]
			// pgx returns various types. For JSON output, we might need to convert some (like []byte)
			if b, ok := val.([]byte); ok {
				row[colName] = string(b)
			} else {
				row[colName] = val
			}
		}
		result = append(result, row)
	}

	jsonResponse(w, map[string]interface{}{
		"rows":            result,
		"substituted_sql": substitutedSQL,
	}, http.StatusOK)
}

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
