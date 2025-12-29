package main

import (
	"encoding/json"
	"net/http"
	"os"

	"mks_sql/pkg/config"
	"mks_sql/pkg/processor"
)

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

func handlePatterns(w http.ResponseWriter, r *http.Request) {
	patterns, err := config.LoadConfigs(nil)
	if err != nil {
		http.Error(w, "Failed to load patterns", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patterns)
}

func handleDoc(w http.ResponseWriter, r *http.Request) {
	// Read reference_guide.md
	// Assuming running from project root
	docPath := "doc/reference_guide.md"
	content, err := os.ReadFile(docPath)
	if err != nil {
		// Try fallback location if running inside mks_so
		docPath = "../doc/reference_guide.md"
		content, err = os.ReadFile(docPath)
		if err != nil {
			// Try fallback if running from cmd/server
			docPath = "../../../doc/reference_guide.md"
			content, err = os.ReadFile(docPath)
			if err != nil {
				http.Error(w, "Documentation not found", http.StatusNotFound)
				return
			}
		}
	}

	// Just return raw markdown for now, frontend can handle or we just show text
	w.Header().Set("Content-Type", "text/markdown")
	w.Write(content)
}

func handleParserRules(w http.ResponseWriter, r *http.Request) {
	docPath := "doc/parser_rules.md"
	content, err := os.ReadFile(docPath)
	if err != nil {
		docPath = "../doc/parser_rules.md"
		content, err = os.ReadFile(docPath)
		if err != nil {
			docPath = "../../../doc/parser_rules.md"
			content, err = os.ReadFile(docPath)
			if err != nil {
				http.Error(w, "Parser rules not found", http.StatusNotFound)
				return
			}
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

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
