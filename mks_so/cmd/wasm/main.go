//go:build js && wasm

package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"syscall/js"

	"mks_sql/pkg/config"
	"mks_sql/pkg/processor"
)

//go:embed config.yaml
var configBytes []byte

func main() {
	// Initialize config for the processor
	// The processor uses config.LoadPatterns() internally, which we need to override or pre-load
	// But processor uses a global variable for compiledPatterns.
	// We might need to expose a way to set patterns in processor or Init it.
	// processor.ProcessSql checks if compiledPatterns is nil.
	// But it calls config.LoadPatterns(nil).
	// We need to inject our byte-loaded patterns.
	// Let's modify processor to allow initializing with bytes or just rely on setting the internal variable?
	// processor.ProcessSql is in package processor.
	// We can't easily access unexported variables.
	// However, we can use config.LoadPatternsWithConfig(configBytes) and then...
	// Wait, processor calls config.LoadPatterns() (no args) if nil.
	// We should probably export a function in processor to Init with patterns.
	// For now, let's assume we can modify processor to check existing logic, or better:
	// We can't change processor.ProcessSql logic easily without editing it.
	// Actually, I can edit processor.go to add Init(patterns).

	c := make(chan struct{}, 0)

	js.Global().Set("processSql", js.FuncOf(processSql))
	js.Global().Set("getPatterns", js.FuncOf(getPatterns))
	js.Global().Set("getTests", js.FuncOf(getTests))
	js.Global().Set("getVersion", js.FuncOf(getVersion))

	fmt.Println("WASM MKS SQL Parser Initialized")
	<-c
}

// Global initialization of regexes for Wasm environment
var wasmPatterns []config.CompiledPattern

func init() {
	wasmPatterns = config.LoadPatternsWithConfig(configBytes)
	// We need to pass these to processor.
	// Since we haven't updated processor to accept patterns injection,
	// We will rely on a new function in processor package: SetPatterns
	processor.SetPatterns(wasmPatterns)
}

func processSql(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return "Error: arguments mismatch"
	}
	sqlText := args[0].String()
	jsonInput := args[1].String()
	minify := false
	if len(args) > 2 {
		minify = args[2].Bool()
	}

	result := processor.ProcessSql(sqlText, jsonInput, minify)
	return result
}

func getPatterns(this js.Value, args []js.Value) interface{} {
	configs, err := config.LoadConfigs(configBytes)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	data, _ := json.Marshal(configs)
	return string(data)
}

func getTests(this js.Value, args []js.Value) interface{} {
	tests := config.LoadTests(configBytes)
	data, _ := json.Marshal(tests)
	return string(data)
}

func getVersion(this js.Value, args []js.Value) interface{} {
	ver, _ := config.LoadBuildInfo(configBytes)
	return ver
}
