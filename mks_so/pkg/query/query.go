package query

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"mks_sql/pkg/config"

	"github.com/jackc/pgx/v5"
)

/*
-------------------------------------------------------
JSON → positional args
-------------------------------------------------------
*/

func JsonParamsToArgs(
	raw json.RawMessage,
	order []string,
) ([]any, error) {

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}

	args := make([]any, len(order))
	for i, key := range order {
		v, ok := m[key]
		if !ok {
			return nil, fmt.Errorf("missing param: %s", key)
		}
		args[i] = v
	}

	return args, nil
}

/*
-------------------------------------------------------
QUERY WITH JSON PARAMS
-------------------------------------------------------
*/

func QueryJSON(
	ctx context.Context,
	conn *pgx.Conn,
	sqlText string,
	jsonParams []byte,
	order []string,
) (pgx.Rows, error) {

	args, err := JsonParamsToArgs(jsonParams, order)
	if err != nil {
		return nil, err
	}

	// Replace $1 with $1::jsonb to help Postgres type inference
	// We do this globally unless it's already part of a cast
	sqlText = strings.ReplaceAll(sqlText, "$1", "$1::jsonb")

	return conn.Query(ctx, sqlText, args...)
}

/*
-------------------------------------------------------
COPY TO STDOUT WITH JSON PARAMS (CSV STREAM)
-------------------------------------------------------
*/

func CopyQueryJSON(
	ctx context.Context,
	conn *pgx.Conn,
	sqlText string,
	jsonParams []byte,
	order []string,
	w io.Writer,
) error {

	processedSQL := sqlText
	// Ensure $1 is cast to jsonb for the COPY command
	processedSQL = strings.ReplaceAll(processedSQL, "$1", "$1::jsonb")

	if len(jsonParams) > 0 {
		args, err := JsonParamsToArgs(jsonParams, order)
		if err != nil {
			return err
		}
		processedSQL = SubstituteParameters(processedSQL, args...)
	}

	csvCfg := config.LoadCSVConfig(nil)
	headerOpt := "FALSE"
	if csvCfg.Header {
		headerOpt = "TRUE"
	}

	copySQL := fmt.Sprintf(`
		COPY (
			%s
		)
		TO STDOUT
		WITH (
			FORMAT %s,
			HEADER %s,
			DELIMITER '%s',
			NULL '%s',
			QUOTE '%s'
		)
	`, processedSQL, csvCfg.Format, headerOpt, csvCfg.Delimiter, csvCfg.Null, csvCfg.Quote)

	_, err := conn.PgConn().CopyTo(ctx, w, copySQL)
	return err
}

// SubstituteParameters mimics PostgreSQL parameter substitution for logging purposes.
// It specifically handles our convention of $1 being a jsonb parameter.
func SubstituteParameters(query string, args ...any) string {
	finalSQL := query
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		var valStr string
		switch v := arg.(type) {
		case string:
			safeStr := strings.ReplaceAll(v, "'", "''")
			// Formatting as jsonb '...' per mks-processing workflow
			valStr = fmt.Sprintf("jsonb '%s'", safeStr)
		case nil:
			valStr = "NULL"
		default:
			// For maps and slices, marshal to JSON string first
			if b, err := json.Marshal(v); err == nil {
				safeStr := strings.ReplaceAll(string(b), "'", "''")
				valStr = fmt.Sprintf("jsonb '%s'", safeStr)
			} else {
				valStr = fmt.Sprintf("%v", v)
			}
		}
		finalSQL = strings.ReplaceAll(finalSQL, placeholder, valStr)
	}
	return finalSQL
}
