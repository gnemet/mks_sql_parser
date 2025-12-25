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

func ProcessSql(sqlText string, jsonInput string) string {
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

	inBlock := false
	blockKeep := true
	minify := false

	if mVal, ok := inputMap["minify"]; ok {
		if mBool, isBool := mVal.(bool); isBool && mBool {
			minify = true
		}
	}

	first := true
	for scanner.Scan() {
		line := scanner.Text()
		lineDeleted := false

		// 1. Block Logic
		for _, cp := range compiledPatterns {
			if cp.Action == "block_start" {
				matches := cp.Re.FindStringSubmatch(line)
				if len(matches) > 0 {
					negate := matches[1] == "!"
					key := matches[2]
					valStr := matches[3]

					cond := evaluator.CheckCondition(inputMap, key, valStr, negate)
					inBlock = true
					blockKeep = cond
				}
			} else if cp.Action == "block_start_jsonpath" {
				matches := cp.Re.FindStringSubmatch(line)
				if len(matches) > 0 {
					exprStr := matches[1]
					cond := evaluator.EvaluateJsonPath(exprStr, inputMap)
					inBlock = true
					blockKeep = cond
				}
			} else if cp.Action == "block_start_nested" {
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
					inBlock = true
					blockKeep = conditionMet
				}
			} else if cp.Action == "block_start_simple_extended" {
				matches := cp.Re.FindStringSubmatch(line)
				if len(matches) > 0 {
					key := matches[1]
					op := matches[2]
					valStr := matches[3]

					cond := evaluator.CheckConditionOp(inputMap, key, valStr, op)
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

		if inBlock && !blockKeep {
			lineDeleted = true
		}

		// 2. Line Logic
		if !lineDeleted {
			shouldDeleteLine := false

			for _, cp := range compiledPatterns {
				if shouldDeleteLine {
					break
				}
				switch cp.Action {
				case "line_filter_jsonpath":
					matches := cp.Re.FindAllStringSubmatch(line, -1)
					for _, m := range matches {
						exprStr := m[1]
						if !evaluator.EvaluateJsonPath(exprStr, inputMap) {
							shouldDeleteLine = true
							break
						}
					}
				case "line_filter_simple_extended":
					matches := cp.Re.FindAllStringSubmatch(line, -1)
					for _, m := range matches {
						key := m[1]
						op := m[2]
						valStr := m[3]
						if !evaluator.CheckConditionOp(inputMap, key, valStr, op) {
							shouldDeleteLine = true
							break
						}
					}
				case "line_filter":
					matches := cp.Re.FindAllStringSubmatch(line, -1)
					for _, m := range matches {
						negateKey := m[1] == "!"
						key := m[2]
						negateVal := m[3] == "!"
						val := m[4]

						var cond bool
						if val != "" {
							cond = evaluator.CheckCondition(inputMap, key, val, negateVal)
						} else {
							cond = evaluator.CheckCondition(inputMap, key, "", negateKey)
						}

						if !cond {
							shouldDeleteLine = true
							break
						}
					}
				case "line_filter_legacy":
					matches := cp.Re.FindAllStringSubmatch(line, -1)
					for _, m := range matches {
						key := m[1]
						if _, ok := inputMap[key]; !ok {
							shouldDeleteLine = true
							break
						}
					}
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
							shouldDeleteLine = true
							break
						}
					}
				case "replace_delete":
					matched := false
					line = cp.Re.ReplaceAllStringFunc(line, func(matchStr string) string {
						sub := cp.Re.FindStringSubmatch(matchStr)
						key := sub[1]
						if val, ok := inputMap[key]; ok {
							matched = true
							return fmt.Sprintf("%v", val)
						} else {
							shouldDeleteLine = true
							return matchStr
						}
					})
					_ = matched
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
			}
			if shouldDeleteLine {
				lineDeleted = true
			}
		}

		if lineDeleted {
			if minify {
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
			} else {
				line = line + " -- kept"
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
