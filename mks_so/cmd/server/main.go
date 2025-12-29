package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API Endpoints
	http.HandleFunc("/process", handleProcess)
	http.HandleFunc("/tests", handleTests)
	http.HandleFunc("/rules", handleRules)
	http.HandleFunc("/doc", handleDoc)
	http.HandleFunc("/parser_rules.md", handleParserRules)
	http.HandleFunc("/version", handleVersion)

	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
