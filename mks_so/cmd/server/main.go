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

	// Wrapper for logging
	logRequest := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[%s] %s %s\n", r.Method, r.URL.Path, r.RemoteAddr)
			handler(w, r)
		}
	}

	// Serve static files with logging
	staticFs := http.FileServer(http.Dir("./cmd/server/static"))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[STATIC] %s %s\n", r.URL.Path, r.RemoteAddr)
		staticFs.ServeHTTP(w, r)
	}))

	// API Endpoints
	http.HandleFunc("/process", logRequest(handleProcess))
	http.HandleFunc("/tests", logRequest(handleTests))
	http.HandleFunc("/rules", logRequest(handleRules))
	http.HandleFunc("/reference_guide.md", logRequest(handleDoc))
	http.HandleFunc("/parser_rules.md", logRequest(handleParserRules))
	http.HandleFunc("/version", logRequest(handleVersion))
	http.HandleFunc("/databases", logRequest(handleDatabases))
	http.HandleFunc("/connect", logRequest(handleConnect))
	http.HandleFunc("/disconnect", logRequest(handleDisconnect))
	http.HandleFunc("/query", logRequest(handleQuery))

	fmt.Printf("Server starting on http://%s:%s\n", host, port)
	if err := http.ListenAndServe(host+":"+port, nil); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
