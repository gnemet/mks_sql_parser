package evaluator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
)

func CheckCondition(inputMap map[string]interface{}, key string, valStr string, negate bool) bool {
	op := "="
	if negate {
		op = "!="
	}
	return CheckConditionOp(inputMap, key, valStr, op)
}

func CheckConditionOp(inputMap map[string]interface{}, key string, valStr string, op string) bool {
	// 1. Check existence
	val, exists := inputMap[key]

	// Existence check only (if valStr empty)
	if valStr == "" && (op == "=" || op == "!=") {
		// If explicit boolean value in input, respect it
		if b, ok := val.(bool); ok {
			if op == "=" {
				return b
			}
			return !b
		} else if bPtr, ok := val.(*bool); ok && bPtr != nil {
			if op == "=" {
				return *bPtr
			}
			return !*bPtr
		}

		if op == "!=" {
			return !exists
		}
		return exists
	}

	if !exists {
		return false
	}

	inputValStr := fmt.Sprintf("%v", val)
	match := false

	switch op {
	case "=":
		match = (inputValStr == valStr)
	case "!=":
		match = (inputValStr != valStr)
	case "~":
		// Regex
		if re, err := regexp.Compile(valStr); err == nil {
			match = re.MatchString(inputValStr)
		}
	case "!~":
		if re, err := regexp.Compile(valStr); err == nil {
			match = !re.MatchString(inputValStr)
		}
	case "~*":
		// IRegex
		if re, err := regexp.Compile("(?i)" + valStr); err == nil {
			match = re.MatchString(inputValStr)
		}
	case "!~*":
		if re, err := regexp.Compile("(?i)" + valStr); err == nil {
			match = !re.MatchString(inputValStr)
		}
	case "~~":
		// LIKE
		pattern := regexFromLike(valStr)
		if re, err := regexp.Compile("^" + pattern + "$"); err == nil {
			match = re.MatchString(inputValStr)
		}
	case "!~~":
		pattern := regexFromLike(valStr)
		if re, err := regexp.Compile("^" + pattern + "$"); err == nil {
			match = !re.MatchString(inputValStr)
		}
	case "~~*":
		// ILIKE
		pattern := regexFromLike(valStr)
		if re, err := regexp.Compile("(?i)^" + pattern + "$"); err == nil {
			match = re.MatchString(inputValStr)
		}
	case "!~~*":
		pattern := regexFromLike(valStr)
		if re, err := regexp.Compile("(?i)^" + pattern + "$"); err == nil {
			match = !re.MatchString(inputValStr)
		}

	case "%", "!%":
		// Similarity (Trigram Jaccard)
		sim := calculateSimilarity(inputValStr, valStr)
		isSim := (sim >= 0.3)
		if op == "%" {
			match = isSim
		} else {
			match = !isSim
		}
	case "%>", "!%>":
		// Word Similarity (Trigram Jaccard with simple sliding window)
		// We want to see if 'valStr' (pattern) is contained in 'inputValStr' (text)
		sim := calculateWordSimilarity(inputValStr, valStr)
		isSim := (sim >= 0.3) // Using 0.3 threshold as per default pg_trgm
		if op == "%>" {
			match = isSim
		} else {
			match = !isSim
		}
	}

	return match
}

func regexFromLike(likeStr string) string {
	s := regexp.QuoteMeta(likeStr)
	s = strings.ReplaceAll(s, "%", ".*")
	s = strings.ReplaceAll(s, "_", ".")
	return s
}

func generateTrigrams(s string) map[string]int {
	// Simple trigram generation: " text " -> "  t", " te", "tex", ...
	// Postgres adds 2 spaces prefix and 1 space suffix.
	// But simply 2 spaces padding front and 1 back is standard pg_trgm.
	// s = "  " + s + " "
	s = "  " + strings.ToLower(s) + " "
	trigrams := make(map[string]int)
	runes := []rune(s)
	if len(runes) < 3 {
		return trigrams
	}
	for i := 0; i <= len(runes)-3; i++ {
		t := string(runes[i : i+3])
		trigrams[t]++
	}
	return trigrams
}

func calculateSimilarity(s1, s2 string) float64 {
	t1 := generateTrigrams(s1)
	t2 := generateTrigrams(s2)

	if len(t1) == 0 && len(t2) == 0 {
		return 1.0 // Both empty/short
	}
	if len(t1) == 0 || len(t2) == 0 {
		return 0.0
	}

	// Jaccard: |Intersection| / |Union|
	// Union = |A| + |B| - |Intersection|
	intersection := 0
	for t, count1 := range t1 {
		if count2, ok := t2[t]; ok {
			// Count intersection?
			// pg_trgm counts MULTIPLICITY?
			// Standard pg_trgm uses set logic mostly but let's be robust.
			// If "nanana" -> "nan": 2.
			// Usually simple set is enough for this use case, but let's do min(count1, count2) if we want multiset.
			// Let's stick to unique trigrams for set Jaccard index as it matches standard definition.
			// "pg_trgm ignores the frequency of trigrams" (documentation says "number of shared trigrams").
			// So Set Intersection.
			_ = count1
			_ = count2
			intersection++
		}
	}

	unique1 := len(t1)
	unique2 := len(t2)
	union := unique1 + unique2 - intersection
	if union == 0 {
		return 1.0
	}
	return float64(intersection) / float64(union)
}

func calculateWordSimilarity(text, pattern string) float64 {
	// pattern in text.
	tPattern := generateTrigrams(pattern)

	if len(tPattern) == 0 {
		return 0.0
	}

	// Ordered trigrams of text for windowing
	s := "  " + strings.ToLower(text) + " "
	runes := []rune(s)
	if len(runes) < 3 {
		return 0.0
	}
	var textTrigrams []string
	for i := 0; i <= len(runes)-3; i++ {
		textTrigrams = append(textTrigrams, string(runes[i:i+3]))
	}

	// Pattern unique trigrams set
	patternSet := make(map[string]bool)
	for t := range tPattern {
		patternSet[t] = true
	}

	maxSim := 0.0

	// Sliding window over textTrigrams.
	// We scan every sub-sequence.
	// Optimization: limit window size?
	// Upper bound of useful window: if window is huge, similarity drops.
	// Minimal window size: 1.
	// But iterating all sub-sequences is O(N^2). len(text) is usually small (<1KB).

	n := len(textTrigrams)
	m := len(patternSet) // size of pattern trigram set

	// Optimization: We only care about windows that might have high similarity.
	// A window should have significant overlap with pattern.
	// Let's just do brute force O(N^2) for now as N is small (column value).

	for i := 0; i < n; i++ {
		// Track intersection for window starting at i
		currentIntersection := 0
		// Use a temp map or re-check?
		// Since we are expanding j, we can update intersection incrementally?
		// But we need to handle duplicates in text window?
		// pg_trgm set based: unique trigrams in window?
		// Yes, "set of trigrams in the extent".

		windowSet := make(map[string]bool)

		for j := i; j < n; j++ {
			t := textTrigrams[j]
			if !windowSet[t] {
				windowSet[t] = true
				if patternSet[t] {
					currentIntersection++
				}
			}

			// Calc Sim
			// intersection of unique trigrams in window AND pattern
			// union = |Pattern| + |Window| - Intersection

			union := m + len(windowSet) - currentIntersection
			sim := float64(currentIntersection) / float64(union)
			if sim > maxSim {
				maxSim = sim
			}
		}
	}

	return maxSim
}

func GetNestedValue(inputMap map[string]interface{}, pathStr string) (bool, interface{}) {
	keys := strings.Split(pathStr, ",")
	current := inputMap

	for i, key := range keys {
		val, ok := current[key]
		if !ok {
			return false, nil
		}
		if i == len(keys)-1 {
			return true, val
		}
		if nextMap, ok := val.(map[string]interface{}); ok {
			current = nextMap
		} else {
			return false, nil
		}
	}
	return true, nil
}

func EvaluateJsonPath(expression string, inputMap map[string]interface{}) bool {
	env := map[string]interface{}{
		"$": inputMap,
	}

	options := []expr.Option{
		expr.Env(env),
		expr.Function("exists", func(params ...interface{}) (interface{}, error) {
			if len(params) == 0 {
				return false, nil
			}
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
		return false
	}

	res, ok := output.(bool)
	if !ok {
		return false
	}
	return res
}
