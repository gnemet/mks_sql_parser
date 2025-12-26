package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"mks_sql/pkg/config"
	"mks_sql/pkg/processor"
)

func TestYamlConfigProcessing(t *testing.T) {
	tests := config.LoadTests(nil)
	if len(tests) == 0 {
		t.Fatalf("No tests found in config.yaml")
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ID_%d", tt.ID), func(t *testing.T) {
			inputBytes, err := json.Marshal(tt.Input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			inputJSON := string(inputBytes)
			// ProcessSql expects the SQL text and the JSON input
			result := processor.ProcessSql(tt.Text, inputJSON, false)

			// Normalize newlines for comparison (optional but good for multi-line)
			// But let's try exact match first as defined in YAML

			// TrimSpace to avoid newline issues
			if strings.TrimSpace(result) != strings.TrimSpace(tt.Expected) {
				// If expected is empty line, maybe result has newline?
				// Visual check in failure message
				t.Errorf("\nID: %d\nINPUT: %s\nTEXT:\n%s\nEXPECTED:\n%q\nGOT:\n%q\n", tt.ID, inputJSON, tt.Text, tt.Expected, result)
			}
		})
	}
}
