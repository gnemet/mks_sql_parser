package processor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"mks_sql/pkg/config"
	"mks_sql/pkg/evaluator"
)

// compiledPatterns is now loaded from package config
var compiledPatterns []config.CompiledPattern

func SetPatterns(patterns []config.CompiledPattern) {
	compiledPatterns = patterns
}

func ProcessSql(sqlText string, jsonInput string, forceMinify bool) string {
	if compiledPatterns == nil {
		compiledPatterns = config.LoadPatterns()
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
		debug := false
		if dVal, ok := inputMap["debug"]; ok {
			if dBool, isBool := dVal.(bool); isBool && dBool {
				debug = true
			}
		}

		if i > 0 {
			sb.WriteString("\n")
		}
		if debug {
			sb.WriteString(fmt.Sprintf("===== %d\n", i))
		}
		sb.WriteString(processSingle(sqlText, inputMap, debug, forceMinify))
	}
	return sb.String()
}

func processSingle(sqlText string, inputMap map[string]interface{}, debug bool, forceMinify bool) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(sqlText))

	inBlock := false
	blockKeep := true
	minify := forceMinify

	if !minify {
		if mVal, ok := inputMap["minify"]; ok {
			if mBool, isBool := mVal.(bool); isBool && mBool {
				minify = true
			}
		}
	}

	first := true
	for scanner.Scan() {
		line := scanner.Text()
		lineDeleted := false

		// 1. Block Logic
		for _, cp := range compiledPatterns {
			matched, keep, newLine := handleBlockLogic(line, inputMap, cp)
			if matched {
				line = newLine
				if cp.Action == "block_end" {
					inBlock = false
					blockKeep = true
				} else {
					inBlock = true
					blockKeep = keep
				}
			}
		}

		if inBlock && !blockKeep {
			lineDeleted = true
		}

		// 2. Line Logic
		if !lineDeleted {
			for _, cp := range compiledPatterns {
				shouldDelete, newLine := handleLineLogic(line, inputMap, cp)
				line = newLine
				if shouldDelete {
					lineDeleted = true
					break
				}
			}
		}

		if lineDeleted {
			if minify || !debug {
				continue
			}
			line = fmt.Sprintf("-- deleted: %s", line)
		} else {
			if minify {
				if idx := strings.Index(line, "--"); idx != -1 {
					line = line[:idx]
				}
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
			} else if debug {
				line = line + " -- kept"
			} else {
				if strings.TrimSpace(line) == "" {
					continue
				}
			}
		}

		if !first {
			result.WriteString("\n")
		}
		result.WriteString(line)
		first = false
	}
	return result.String()
}

func handleBlockLogic(line string, inputMap map[string]interface{}, cp config.CompiledPattern) (bool, bool, string) {
	switch cp.Action {
	case "block_start":
		matches := cp.Re.FindStringSubmatch(line)
		if len(matches) > 0 {
			negate := matches[1] == "!"
			key := matches[2]
			valStr := matches[3]
			return true, evaluator.CheckCondition(inputMap, key, valStr, negate), cp.Re.ReplaceAllString(line, "")
		}
	case "block_start_jsonpath":
		matches := cp.Re.FindStringSubmatch(line)
		if len(matches) > 0 {
			exprStr := matches[1]
			return true, evaluator.EvaluateJsonPath(exprStr, inputMap), cp.Re.ReplaceAllString(line, "")
		}
	case "block_start_nested":
		matches := cp.Re.FindStringSubmatch(line)
		if len(matches) > 0 {
			pathStr := matches[1]
			valStr := matches[2]
			exists, val := evaluator.GetNestedValue(inputMap, pathStr)
			conditionMet := exists
			if valStr != "" {
				if !exists {
					conditionMet = false
				} else {
					conditionMet = (fmt.Sprintf("%v", val) == valStr)
				}
			}
			return true, conditionMet, cp.Re.ReplaceAllString(line, "")
		}
	case "block_start_simple_extended":
		matches := cp.Re.FindStringSubmatch(line)
		if len(matches) > 0 {
			key := matches[1]
			op := matches[2]
			valStr := matches[3]
			return true, evaluator.CheckConditionOp(inputMap, key, valStr, op), cp.Re.ReplaceAllString(line, "")
		}
	case "block_end":
		if cp.Re.MatchString(line) {
			return true, true, cp.Re.ReplaceAllString(line, "")
		}
	}
	return false, false, line
}

func handleLineLogic(line string, inputMap map[string]interface{}, cp config.CompiledPattern) (bool, string) {
	shouldDelete := false
	switch cp.Action {
	case "line_filter_jsonpath":
		matches := cp.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if !evaluator.EvaluateJsonPath(m[1], inputMap) {
				shouldDelete = true
			}
		}
		line = cp.Re.ReplaceAllString(line, "")
	case "line_filter_simple_extended":
		matches := cp.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if !evaluator.CheckConditionOp(inputMap, m[1], m[3], m[2]) {
				shouldDelete = true
			}
		}
		line = cp.Re.ReplaceAllString(line, "")
	case "line_filter":
		matches := cp.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			negateKey := m[1] == "!"
			key := m[2]
			negateVal := m[3] == "!"
			val := m[4]
			if val != "" {
				if !evaluator.CheckCondition(inputMap, key, val, negateVal) {
					shouldDelete = true
				}
			} else {
				if !evaluator.CheckCondition(inputMap, key, "", negateKey) {
					shouldDelete = true
				}
			}
		}
		line = cp.Re.ReplaceAllString(line, "")
	case "line_filter_legacy":
		matches := cp.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if _, ok := inputMap[m[1]]; !ok {
				shouldDelete = true
			}
		}
		line = cp.Re.ReplaceAllString(line, "")
	case "line_filter_nested":
		matches := cp.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			pathStr := m[1]
			valStr := ""
			if len(m) > 2 {
				valStr = m[2]
			}
			exists, val := evaluator.GetNestedValue(inputMap, pathStr)
			conditionMet := exists
			if valStr != "" {
				if !exists {
					conditionMet = false
				} else {
					conditionMet = (fmt.Sprintf("%v", val) == valStr)
				}
			}
			if !conditionMet {
				shouldDelete = true
			}
		}
		line = cp.Re.ReplaceAllString(line, "")
	case "replace_delete":
		line = cp.Re.ReplaceAllStringFunc(line, func(matchStr string) string {
			sub := cp.Re.FindStringSubmatch(matchStr)
			key := sub[1]
			if val, ok := inputMap[key]; ok {
				return fmt.Sprintf("%v", val)
			} else {
				shouldDelete = true
				return matchStr
			}
		})
	case "replace_empty":
		line = cp.Re.ReplaceAllStringFunc(line, func(matchStr string) string {
			sub := cp.Re.FindStringSubmatch(matchStr)
			key := sub[1]
			if val, ok := inputMap[key]; ok {
				return fmt.Sprintf("%v", val)
			} else {
				return ""
			}
		})
	}
	return shouldDelete, line
}
