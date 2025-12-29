package processor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"mks_sql/pkg/config"
	"mks_sql/pkg/evaluator"
)

// compiledRules is now loaded from package config
var compiledRules []config.CompiledRule

func SetRules(rules []config.CompiledRule) {
	compiledRules = rules
}

func ProcessSql(sqlText string, jsonInput string, forceMinify bool) string {
	if compiledRules == nil {
		compiledRules = config.LoadRules()
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
		originalLine := line
		lineDeleted := false

		// 1. Block Logic
		for _, rule := range compiledRules {
			matched, keep, newLine := handleBlockLogic(line, inputMap, rule)
			if matched {
				line = newLine
				if rule.Action == "block_end" {
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
			for _, rule := range compiledRules {
				shouldDelete, newLine := handleLineLogic(line, inputMap, rule)
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
			line = fmt.Sprintf("-- deleted: %s", originalLine)
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
				line = fmt.Sprintf("%s -- kept [original: %s]", line, originalLine)
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

func handleBlockLogic(line string, inputMap map[string]interface{}, rule config.CompiledRule) (bool, bool, string) {
	switch rule.Action {
	case "block_start":
		matches := rule.Re.FindStringSubmatch(line)
		if len(matches) > 0 {
			negateKey := matches[1] == "!"
			key := matches[2]
			op := matches[3]
			valStr := matches[4]

			negate := negateKey
			if op == "!:" {
				negate = true
			}
			// Special handling: if op is empty, valStr is empty.
			// If op is : or !:, valStr is the value.

			return true, evaluator.CheckCondition(inputMap, key, valStr, negate), rule.Re.ReplaceAllString(line, "")
		}
	case "block_start_jsonpath":
		matches := rule.Re.FindStringSubmatch(line)
		if len(matches) > 0 {
			exprStr := matches[1]
			return true, evaluator.EvaluateJsonPath(exprStr, inputMap), rule.Re.ReplaceAllString(line, "")
		}
	case "block_start_nested":
		matches := rule.Re.FindStringSubmatch(line)
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
			return true, conditionMet, rule.Re.ReplaceAllString(line, "")
		}
	case "block_start_simple_extended":
		matches := rule.Re.FindStringSubmatch(line)
		if len(matches) > 0 {
			key := matches[1]
			op := matches[2]
			valStr := matches[3]
			return true, evaluator.CheckConditionOp(inputMap, key, valStr, op), rule.Re.ReplaceAllString(line, "")
		}
	case "block_end":
		if rule.Re.MatchString(line) {
			return true, true, rule.Re.ReplaceAllString(line, "")
		}
	}
	return false, false, line
}

func handleLineLogic(line string, inputMap map[string]interface{}, rule config.CompiledRule) (bool, string) {
	shouldDelete := false
	switch rule.Action {
	case "line_filter_jsonpath":
		matches := rule.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if !evaluator.EvaluateJsonPath(m[1], inputMap) {
				shouldDelete = true
			}
		}
	case "line_filter_simple_extended":
		matches := rule.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if !evaluator.CheckConditionOp(inputMap, m[1], m[3], m[2]) {
				shouldDelete = true
			}
		}
	case "line_filter":
		matches := rule.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			negateKey := m[1] == "!"
			key := m[2]
			op := m[3]
			val := m[4]

			negateVal := op == "!:"

			if val != "" || op != "" {
				// We have a value check (key:value or key!:value or key:!value handled by regex?)
				// Regex matches (:|!:) so we know if it's !=

				// evaluator.CheckCondition(..., negate) returns:
				// if negate is false: key exists and equals val
				// if negate is true: key exists and NOT equals val (or key missing?)

				// Wait, CheckCondition logic:
				// if val != "" -> if negate: return valStr != val

				conditionMet := evaluator.CheckCondition(inputMap, key, val, false)
				if negateVal || negateKey {
					// Logic:
					// #!key:val -> Delete if key=val (negate match)
					// #key!:val -> Delete if key=val (negate match) ? No.
					// #key!:val -> Keep if key!=val. Delete if key==val.

					// If conditionMet (key==val) is true, and we want !:, then we delete.
					if conditionMet {
						shouldDelete = true
					}
				} else {
					// #key:val -> Keep if key=val. Delete if key!=val or missing.
					if !conditionMet {
						shouldDelete = true
					}
				}
			} else {
				// Existence check only: #key or #!key
				if !evaluator.CheckCondition(inputMap, key, "", negateKey) {
					shouldDelete = true
				}
			}
		}
	case "line_filter_legacy":
		matches := rule.Re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if _, ok := inputMap[m[1]]; !ok {
				shouldDelete = true
			}
		}
		line = rule.Re.ReplaceAllString(line, "")
	case "line_filter_nested":
		matches := rule.Re.FindAllStringSubmatch(line, -1)
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
	case "replace_delete":
		line = rule.Re.ReplaceAllStringFunc(line, func(matchStr string) string {
			sub := rule.Re.FindStringSubmatch(matchStr)
			key := sub[1]
			if val, ok := inputMap[key]; ok {
				return fmt.Sprintf("%v", val)
			} else {
				shouldDelete = true
				return matchStr
			}
		})
	case "replace_empty":
		line = rule.Re.ReplaceAllStringFunc(line, func(matchStr string) string {
			sub := rule.Re.FindStringSubmatch(matchStr)
			key := sub[1]
			if val, ok := inputMap[key]; ok {
				return fmt.Sprintf("%v", val)
			} else {
				return ""
			}
		})
	case "delete":
		line = rule.Re.ReplaceAllStringFunc(line, func(matchStr string) string {
			sub := rule.Re.FindStringSubmatch(matchStr)
			if len(sub) > 1 {
				key := sub[1]
				// Check for nested path syntax: {key1,key2}
				if strings.HasPrefix(key, "{") && strings.HasSuffix(key, "}") {
					pathStr := key[1 : len(key)-1]
					exists, _ := evaluator.GetNestedValue(inputMap, pathStr)
					if exists {
						return matchStr
					}
				} else {
					// Check for exact match
					if _, ok := inputMap[key]; ok {
						return matchStr
					}
					// Check for case-insensitive match
					for k := range inputMap {
						if strings.EqualFold(k, key) {
							return matchStr
						}
					}
				}
			}
			shouldDelete = true
			return matchStr
		})
	}
	return shouldDelete, line
}
