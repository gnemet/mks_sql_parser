package main

import (
	"fmt"
	"net/http"

	"mks_sql/pkg/config"
)

func main() {
	appCfg := config.LoadAppConfig(nil)
	port := fmt.Sprintf("%d", appCfg.Port)
	if port == "0" {
		port = "8080"
	}
	host := appCfg.Host
	if host == "" {
		host = "localhost"
	}

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API Endpoints
	http.HandleFunc("/process", handleProcess)
	http.HandleFunc("/tests", handleTests)
	http.HandleFunc("/rules", handleRules)
	http.HandleFunc("/reference_guide.md", handleDoc)
	http.HandleFunc("/parser_rules.md", handleParserRules)
	http.HandleFunc("/version", handleVersion)
	http.HandleFunc("/databases", handleDatabases)
	http.HandleFunc("/connect", handleConnect)
	http.HandleFunc("/disconnect", handleDisconnect)
	http.HandleFunc("/query", handleQuery)

	fmt.Printf("Server starting on http://%s:%s\n", host, port)
	if err := http.ListenAndServe(host+":"+port, nil); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
