package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"mks_sql/pkg/config"
	"mks_sql/pkg/processor"
	"mks_sql/pkg/query"

	"github.com/jackc/pgx/v5"
)

func TestDatabaseCSVOutput(t *testing.T) {
	// 1. Load configuration
	tests := config.LoadTests(nil)
	dbs := config.LoadDatabaseConfigs(nil)
	if len(dbs) == 0 {
		t.Skip("No database configurations found, skipping database test")
	}

	// Use connection info from config.yaml as requested
	var dbCfg config.DatabaseConfig
	found := false
	for _, db := range dbs {
		if db.Name == "zenbook" {
			dbCfg = db
			found = true
			break
		}
	}
	if !found {
		dbCfg = dbs[0]
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.Database, dbCfg.Schema)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// 2. Iterate through tests with csv_expected
	for _, tt := range tests {
		if tt.CsvExpected == "" {
			continue
		}

		t.Run(fmt.Sprintf("DB_CSV_ID_%d", tt.ID), func(t *testing.T) {
			if !tt.Passed {
				t.Skipf("Skipping test ID %d because passed=false", tt.ID)
			}

			inputBytes, _ := json.Marshal(tt.Input)
			inputJSON := string(inputBytes)

			// Process the SQL using our MKS rules
			processedSQL := processor.ProcessSql(tt.Text, inputJSON, false)

			// Prepare for CSV output
			var buf bytes.Buffer

			wrappedParams := map[string]interface{}{"data": tt.Input}
			wrappedBytes, _ := json.Marshal(wrappedParams)
			wrappedOrder := []string{"data"}

			// Execute using COPY mode
			err := query.CopyQueryJSON(ctx, conn, processedSQL, wrappedBytes, wrappedOrder, &buf)
			if err != nil {
				t.Fatalf("CopyQueryJSON failed: %v", err)
			}

			got := buf.String()

			// Print data to terminal as requested
			fmt.Printf("\n--- [START] TEST ID %d CSV DATA ---\n%s--- [END] ---\n", tt.ID, got)

			lines := strings.Split(strings.TrimSpace(got), "\n")
			var dataLines []string
			if len(lines) > 1 {
				dataLines = lines[1:] // Skip header
			} else {
				dataLines = lines
			}

			dataResult := strings.Join(dataLines, "\n")

			// Normalization function
			normalize := func(s string) string {
				s = strings.ReplaceAll(s, ";", ",")
				s = strings.ReplaceAll(s, "true", "t")
				s = strings.ReplaceAll(s, "false", "f")
				return strings.TrimSpace(s)
			}

			normalizedGot := normalize(dataResult)
			normalizedExpected := normalize(tt.CsvExpected)

			if normalizedGot != normalizedExpected {
				t.Errorf("\nID: %d\nGOT (normalized):\n%q\nEXPECTED (normalized):\n%q\n", tt.ID, normalizedGot, normalizedExpected)
			}
		})
	}
}
