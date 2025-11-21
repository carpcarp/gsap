package sap

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// EnumCoercer handles coercion to enum types
type EnumCoercer struct{}

// CoerceToEnum attempts to coerce a value to an enum type
func CoerceToEnum(value interface{}, enumType reflect.Type, score *Score) (interface{}, error) {
	stringVal, err := coerceValueToString(value)
	if err != nil {
		return nil, err
	}

	// Get all possible enum values
	enumValues := getEnumValues(enumType)

	// Try exact match first
	for _, ev := range enumValues {
		if ev == stringVal {
			return stringVal, nil
		}
	}

	// Try case-insensitive match
	lowerVal := strings.ToLower(stringVal)
	for _, ev := range enumValues {
		if strings.ToLower(ev) == lowerVal {
			score.AddFlag("EnumCaseInsensitive", 1)
			return ev, nil
		}
	}

	// Try fuzzy match with Unicode normalization
	bestMatch := fuzzyMatchEnum(stringVal, enumValues)
	if bestMatch != "" {
		score.AddFlag("EnumFuzzyMatch", 2)
		return bestMatch, nil
	}

	// If still no match, return the original string
	// (Go's json.Unmarshal will handle the actual type assertion)
	return stringVal, nil
}

// coerceValueToString is a helper to convert any value to string
func coerceValueToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v)), nil
		}
		return fmt.Sprintf("%f", v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("cannot convert %T to string", value)
	}
}

// getEnumValues extracts possible enum values
// For Go types, we need reflection to find constants
func getEnumValues(enumType reflect.Type) []string {
	// In Go, enums are typically string constants
	// This is a simplified version - in reality you'd want to use
	// struct tags or a registry to define enum values

	var values []string
	// For now, return empty - the caller should populate this
	return values
}

// fuzzyMatchEnum attempts to match a string to enum values with fuzzy matching
func fuzzyMatchEnum(input string, enumValues []string) string {
	type scoreResult struct {
		value string
		score int
	}

	var results []scoreResult

	for _, enumValue := range enumValues {
		score := stringDistance(input, enumValue)
		results = append(results, scoreResult{enumValue, score})
	}

	// Return the best match if it's close enough
	if len(results) > 0 {
		best := results[0]
		for _, r := range results {
			if r.score < best.score {
				best = r
			}
		}

		// Only return if reasonably close match (< 50% different)
		threshold := (len(input) + len(best.value)) / 2
		if best.score <= threshold {
			return best.value
		}
	}

	return ""
}

// stringDistance calculates Levenshtein distance with normalization
func stringDistance(s1, s2 string) int {
	// First try with accent normalization
	s1Norm := normalizeString(s1)
	s2Norm := normalizeString(s2)

	return levenshteinDistance(s1Norm, s2Norm)
}

// normalizeString normalizes a string for comparison
func normalizeString(s string) string {
	// Remove accents and convert to lowercase
	s = strings.ToLower(s)

	// Simple accent removal map
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'ä': 'a', 'â': 'a', 'ã': 'a', 'å': 'a',
		'é': 'e', 'è': 'e', 'ë': 'e', 'ê': 'e',
		'í': 'i', 'ì': 'i', 'ï': 'i', 'î': 'i',
		'ó': 'o', 'ò': 'o', 'ö': 'o', 'ô': 'o', 'õ': 'o',
		'ú': 'u', 'ù': 'u', 'ü': 'u', 'û': 'u',
		'ý': 'y', 'ỳ': 'y', 'ÿ': 'y',
		'ç': 'c', 'č': 'c',
		'ñ': 'n',
		'ß': 's',
		'æ': 'a', 'œ': 'o',
	}

	var result strings.Builder
	for _, ch := range s {
		if replacement, ok := replacements[ch]; ok {
			result.WriteRune(replacement)
		} else if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == ' ' {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	runes1 := []rune(s1)
	runes2 := []rune(s2)

	len1 := len(runes1)
	len2 := len(runes2)

	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	// Create a DP table
	dp := make([][]int, len1+1)
	for i := range dp {
		dp[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		dp[0][j] = j
	}

	// Fill the DP table
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if runes1[i-1] != runes2[j-1] {
				cost = 1
			}

			dp[i][j] = min(
				dp[i-1][j]+1,      // deletion
				dp[i][j-1]+1,      // insertion
				dp[i-1][j-1]+cost, // substitution
			)
		}
	}

	return dp[len1][len2]
}

func min(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values {
		if v < m {
			m = v
		}
	}
	return m
}
