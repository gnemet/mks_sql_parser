package processor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/spf13/viper"
)

type PatternConfig struct {
	Regex  string `mapstructure:"regex"`
	Action string `mapstructure:"action"`
}

var compiledPatterns []struct {
	Re     *regexp.Regexp
	Action string
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../..")
	viper.AddConfigPath("..")

	if err := viper.ReadInConfig(); err != nil {
		// defaults if missing
	}

	var configs []PatternConfig
	viper.UnmarshalKey("patterns", &configs)

	if len(configs) == 0 {
		// Fallbacks
		configs = []PatternConfig{
			{Regex: `--\s*<\{\s*(.*?)\s*\}`, Action: "block_start_jsonpath"},
			{Regex: `--\s*#\{\s*(.*?)\s*\}`, Action: "line_filter_jsonpath"},
			{Regex: `\$1->[$>^&@]{0,1}'([^']+)'`, Action: "line_filter_legacy"},
			{Regex: `\B[\$:]([a-zA-Z_]\w*)`, Action: "replace_delete"},
		}
	}

	compiledPatterns = nil
	for _, cfg := range configs {
		re, err := regexp.Compile(cfg.Regex)
		if err != nil {
			fmt.Printf("Error compiling regex '%s': %s\n", cfg.Regex, err)
			continue
		}
		compiledPatterns = append(compiledPatterns, struct {
			Re     *regexp.Regexp
			Action string
		}{Re: re, Action: cfg.Action})
	}
}

func ProcessSql(sqlText string, jsonInput string) string {
	if compiledPatterns == nil {
		InitConfig()
	}

	var inputList []map[string]interface{}
	if jsonInput != "" {
		if err := json.Unmarshal([]byte(jsonInput), &inputList); err != nil {
			var singleInput map[string]interface{}
			if err2 := json.Unmarshal([]byte(jsonInput), &singleInput); err2 == nil {
				inputList = append(inputList, singleInput)
			}
		}
	} else {
		inputList = append(inputList, make(map[string]interface{}))
	}

	if len(inputList) == 0 && jsonInput != "" {
		inputList = append(inputList, make(map[string]interface{}))
	}

	var sb strings.Builder
	for i, inputMap := range inputList {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("===== %d\n", i))
		sb.WriteString(processSingle(sqlText, inputMap))
	}
	return sb.String()
}

func processSingle(sqlText string, inputMap map[string]interface{}) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(sqlText))

	// Block state
	inBlock := false
	blockKeep := true // If inBlock is true, this determines if we keep lines

	first := true
	for scanner.Scan() {
		line := scanner.Text()

		// status for this line
		lineDeleted := false

		// If we are in a block that should be skipped, we mark line deleted by default (unless it's the end tag?)
		// Usually block markers are comments, so we can process them even if "skipping".
		// But if we are skipping, we effectively comment out the line.

		// 1. Process Block Markers logic first (can change state)
		for _, cp := range compiledPatterns {
			if cp.Action == "block_start" {
				matches := cp.Re.FindStringSubmatch(line)
				if len(matches) > 0 {
					// Found block start
					// Groups: 0 full, 1 "!" (optional), 2 key, 3 value (optional)
					negate := matches[1] == "!"
					key := matches[2]
					valStr := matches[3]

					cond := checkCondition(inputMap, key, valStr, negate)

					// Flatten: we don't support nesting so just set state
					inBlock = true
					blockKeep = cond
				}
			} else if cp.Action == "block_start_jsonpath" {
				matches := cp.Re.FindStringSubmatch(line)
				if len(matches) > 0 {
					// Group 1 is the expression
					exprStr := matches[1]
					cond := evaluateJsonPath(exprStr, inputMap)
					inBlock = true
					blockKeep = cond
				}
			} else if cp.Action == "block_end" {
				if cp.Re.MatchString(line) {
					inBlock = false
					blockKeep = true
				}
			}
		}

		// If we are inside a skipped block, this line is deleted
		if inBlock && !blockKeep {
			lineDeleted = true
		}

		// 2. Line Filters and Replacements (only if not already deleted)
		// Note: Even if lineDeleted is true (due to block), we might want to preserve the marker line itself?
		// User requirements implied content is removed. We will convert to "-- deleted: ..."

		if !lineDeleted {
			shouldDeleteLine := false

			for _, cp := range compiledPatterns {
				if shouldDeleteLine {
					break
				}

				switch cp.Action {
				case "line_filter_jsonpath":
					// #{ expression }
					matches := cp.Re.FindAllStringSubmatch(line, -1)
					for _, m := range matches {
						exprStr := m[1]
						if !evaluateJsonPath(exprStr, inputMap) {
							shouldDeleteLine = true
							break
						}
					}
				case "line_filter":
					// #(!?)(\w+)(?::(!?)(\w+))?
					// Groups: 1 !, 2 key, 3 val!, 4 val
					matches := cp.Re.FindAllStringSubmatch(line, -1)
					for _, m := range matches {
						negateKey := m[1] == "!"
						key := m[2]
						negateVal := m[3] == "!"
						val := m[4]

						// Check logic:
						// If val is empty, it's existence check on key.
						// If val is present, it's value check.
						// Wait, regex might match #key as G1="", G2="key", G3="", G4=""
						// #!key -> G1="!", G2="key"
						// #key:val -> G2="key", G4="val"

						// To properly handle #key:!val vs #key:val
						// We need to implement checkCondition helper properly

						// checkCondition handles (key, val, negate)
						// line filter logic: if condition NOT met, delete line.

						var cond bool
						if val != "" {
							// Value check
							cond = checkCondition(inputMap, key, val, negateVal)
							// Note: negateKey would mean "NOT (key:val)"?
							// Usually syntax is #!key (not exists) OR #key:!val (val not equals)
							// If regex is #(!?)key... then negation applies to key existence usually.
							// Let's assume ! applies to the whole check if at start, or val if at val.
						} else {
							// Existence check
							cond = checkCondition(inputMap, key, "", negateKey)
						}

						if !cond {
							shouldDeleteLine = true
							break
						}
					}

				case "line_filter_legacy":
					// $1->'key'
					// Group 1 is key.
					matches := cp.Re.FindAllStringSubmatch(line, -1)
					for _, m := range matches {
						key := m[1]
						// Legacy: match if key exists
						if _, ok := inputMap[key]; !ok {
							shouldDeleteLine = true
							break
						}
					}

				case "replace_delete":
					// :key or $key
					// Group 1 is key
					matched := false
					line = cp.Re.ReplaceAllStringFunc(line, func(matchStr string) string {
						sub := cp.Re.FindStringSubmatch(matchStr)
						key := sub[1]
						if val, ok := inputMap[key]; ok {
							matched = true
							return fmt.Sprintf("%v", val)
						} else {
							shouldDeleteLine = true
							return matchStr // will be deleted anyway
						}
					})
					_ = matched // unused

				case "replace_empty":
					// %key%
					// Group 1 is key
					line = cp.Re.ReplaceAllStringFunc(line, func(matchStr string) string {
						sub := cp.Re.FindStringSubmatch(matchStr)
						key := sub[1]
						if val, ok := inputMap[key]; ok {
							return fmt.Sprintf("%v", val)
						} else {
							return "" // replace with empty/space
						}
					})
				}
			}

			if shouldDeleteLine {
				lineDeleted = true
			}
		}

		if lineDeleted {
			line = fmt.Sprintf("-- deleted: %s", line)
		} else {
			// Keep
			// Only append " -- kept" if it was a conditional line?
			// User examples showed "-- kept" on lines that had checks.
			// Simple approach: append kept to everything valid?
			// Or try to detect if we did something?
			// Existing logic appended it. Let's append if it contains markers or replacements?
			// "if key exists then keep the line , mark it comment"
			// Let's stick to appending " -- kept" for now as it aids debugging/verification.
			line = line + " -- kept"
		}

		if !first {
			result.WriteString("\n")
		}
		result.WriteString(line)
		first = false
	}
	return result.String()
}

func checkCondition(inputMap map[string]interface{}, key string, valStr string, negate bool) bool {
	// 1. Check existence
	val, exists := inputMap[key]

	// Existence check only (if valStr empty)
	if valStr == "" {
		if negate {
			return !exists
		}
		return exists
	}

	// Value check
	// If key missing, condition fails immediately?
	// "key:val" -> implies key must exist AND equal val
	if !exists {
		// If we wanted check for "key:val", and key missing, it's False.
		// If we wanted "key:!val", and key missing -> is it True (not equal) or False (key must exist)?
		// Usually strict: key must exist.
		return false
	}

	// Compare values
	// define strict equality for strings, convert others
	inputValStr := fmt.Sprintf("%v", val)

	match := (inputValStr == valStr)

	if negate {
		return !match
	}
	return match
}

func evaluateJsonPath(expression string, inputMap map[string]interface{}) bool {
	// Env wrapping
	env := map[string]interface{}{
		"$": inputMap,
	}

	options := []expr.Option{
		expr.Env(env),
		// Register "exists" function
		expr.Function("exists", func(params ...interface{}) (interface{}, error) {
			if len(params) == 0 {
				return false, nil
			}
			// If param is nil, key is missing
			return params[0] != nil, nil
		}),
	}

	program, err := expr.Compile(expression, options...)
	if err != nil {
		fmt.Printf("Error compiling expression '%s': %v\n", expression, err)
		return false
	}

	output, err := expr.Run(program, env)
	if err != nil {
		// Runtime error usually means failure of condition
		return false
	}

	res, ok := output.(bool)
	if !ok {
		return false
	}
	return res
}
