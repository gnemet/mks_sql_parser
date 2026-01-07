package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"mks_sql/pkg/config"
)

func main() {
	destDir := flag.String("dest", "cmd/server/static", "Destination directory")
	flag.Parse()

	if err := os.MkdirAll(*destDir, 0755); err != nil {
		log.Fatalf("Failed to create destination directory: %v", err)
	}

	// 1. Databases
	dbs := config.LoadDatabaseConfigs(nil)
	saveJSON(filepath.Join(*destDir, "databases.json"), dbs)

	// 2. Tests
	tests := config.LoadTests(nil)
	saveJSON(filepath.Join(*destDir, "tests.json"), tests)

	// 3. Rules
	rules, err := config.LoadConfigs(nil)
	if err != nil {
		log.Printf("Warning: failed to load rules: %v", err)
	} else {
		saveJSON(filepath.Join(*destDir, "rules.json"), rules)
	}

	// 4. Version and app info
	version, lastBuild := config.LoadBuildInfo(nil)
	appCfg := config.LoadAppConfig(nil)
	info := map[string]interface{}{
		"version":    version,
		"last_build": lastBuild,
		"sql_limits": appCfg.SqlLimits,
	}
	saveJSON(filepath.Join(*destDir, "version.json"), info)

	fmt.Println("Static data generation complete.")
}

func saveJSON(path string, data interface{}) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to encode JSON to %s: %v", path, err)
	}
	fmt.Printf("Generated %s\n", path)
}
