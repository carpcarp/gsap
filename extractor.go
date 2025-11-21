package sap

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// NewExtractor creates a new JSON extractor
func NewExtractor(opts *ParseOptions) *Extractor {
	return &Extractor{
		parser: &FixingParser{
			allowIncomplete: opts.Streaming.AllowIncompleteJSON,
		},
	}
}

// ExtractJSON extracts potential JSON from text
// Returns candidates in order of likelihood
func (e *Extractor) ExtractJSON(input string) ([]JSONCandidate, error) {
	var candidates []JSONCandidate

	// First, try standard JSON parsing (most likely to succeed)
	trimmed := strings.TrimSpace(input)
	if isValidJSON(trimmed) {
		candidates = append(candidates, JSONCandidate{
			JSON:  trimmed,
			Index: strings.Index(input, trimmed),
		})
		return candidates, nil
	}

	// Second, try markdown code blocks
	markdownCandidates := e.extractMarkdownJSON(input)
	candidates = append(candidates, markdownCandidates...)

	// Third, try finding all JSON objects/arrays in text
	naiveJSONs := e.findJSONInText(input)
	candidates = append(candidates, naiveJSONs...)

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no JSON found in input")
	}

	return candidates, nil
}

// extractMarkdownJSON extracts JSON from markdown code blocks
func (e *Extractor) extractMarkdownJSON(input string) []JSONCandidate {
	var candidates []JSONCandidate

	// Match ```json ... ``` or ``` ... ```
	re := regexp.MustCompile("```(?:json|JSON)?\\s*\\n([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatchIndex(input, -1)

	for _, match := range matches {
		// match[2] is start of group 1, match[3] is end
		json := input[match[2]:match[3]]
		json = strings.TrimSpace(json)

		if json != "" {
			candidates = append(candidates, JSONCandidate{
				JSON:  json,
				Index: match[2],
			})
		}
	}

	return candidates
}

// findJSONInText finds JSON objects/arrays in text
// This handles cases where JSON is embedded in natural language
func (e *Extractor) findJSONInText(input string) []JSONCandidate {
	var candidates []JSONCandidate

	// Try to find JSON objects { ... }
	candidates = append(candidates, e.findJSONBlocks(input, '{', '}')...)

	// Try to find JSON arrays [ ... ]
	candidates = append(candidates, e.findJSONBlocks(input, '[', ']')...)

	return candidates
}

// findJSONBlocks finds balanced braces/brackets in text
func (e *Extractor) findJSONBlocks(input string, openChar, closeChar rune) []JSONCandidate {
	var candidates []JSONCandidate
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		if runes[i] == openChar {
			// Try to find matching close character
			depth := 1
			inString := false
			escaped := false

			j := i + 1
			for j < len(runes) && depth > 0 {
				ch := runes[j]

				if escaped {
					escaped = false
					j++
					continue
				}

				if ch == '\\' {
					escaped = true
					j++
					continue
				}

				if ch == '"' && !escaped {
					inString = !inString
					j++
					continue
				}

				if !inString {
					if ch == openChar {
						depth++
					} else if ch == closeChar {
						depth--
					}
				}

				j++
			}

			// If we found a match, extract it
			if depth == 0 {
				json := string(runes[i : j])
				candidates = append(candidates, JSONCandidate{
					JSON:  json,
					Index: i,
				})
			}
		}
	}

	return candidates
}

// isValidJSON checks if a string is valid JSON
func isValidJSON(input string) bool {
	var v interface{}
	return json.Unmarshal([]byte(input), &v) == nil
}
