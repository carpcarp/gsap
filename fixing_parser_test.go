package sap

import (
	"encoding/json"
	"strings"
	"testing"
)

// mustBeValidJSON is a test helper that unmarshals the result into a generic
// value and fails the test if the JSON is malformed.
func mustBeValidJSON(t *testing.T, s string) {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("result is not valid JSON: %v\nraw: %s", err, s)
	}
}

// mustUnmarshalObject unmarshals s into a map and fails the test on error.
func mustUnmarshalObject(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("result is not a valid JSON object: %v\nraw: %s", err, s)
	}
	return m
}

// mustUnmarshalArray unmarshals s into a slice and fails the test on error.
func mustUnmarshalArray(t *testing.T, s string) []interface{} {
	t.Helper()
	var a []interface{}
	if err := json.Unmarshal([]byte(s), &a); err != nil {
		t.Fatalf("result is not a valid JSON array: %v\nraw: %s", err, s)
	}
	return a
}

// ---------------------------------------------------------------------------
// closeUnclosedStructures — exercised through FixJSON with truncated inputs
// ---------------------------------------------------------------------------

func TestFixJSON_CloseUnclosedStructures(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKeys  []string   // expected top-level keys (nil to skip check)
		wantArray bool       // true if the result should be an array
		wantLen   int        // expected array length (-1 to skip)
	}{
		{
			name:     "unclosed object",
			input:    `{"name": "Alice"`,
			wantKeys: []string{"name"},
		},
		{
			name:      "unclosed array",
			input:     `["a", "b"`,
			wantArray: true,
			wantLen:   2,
		},
		{
			name:     "nested unclosed array inside object",
			input:    `{"items": ["a", "b"`,
			wantKeys: []string{"items"},
		},
		{
			name:     "unclosed object with trailing comma removed",
			input:    `{"a": 1, "b": 2,`,
			wantKeys: []string{"a", "b"},
		},
		{
			name:     "multiple nested unclosed structures",
			input:    `{"a": {"b": [1, 2`,
			wantKeys: []string{"a"},
		},
		{
			name:     "deeply nested all unclosed",
			input:    `{"l1": {"l2": {"l3": [true`,
			wantKeys: []string{"l1"},
		},
		{
			name:      "unclosed array with trailing comma",
			input:     `[1, 2, 3,`,
			wantArray: true,
			wantLen:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			if tt.wantArray {
				a := mustUnmarshalArray(t, got)
				if tt.wantLen >= 0 && len(a) != tt.wantLen {
					t.Errorf("expected array length %d, got %d", tt.wantLen, len(a))
				}
			} else if tt.wantKeys != nil {
				m := mustUnmarshalObject(t, got)
				for _, k := range tt.wantKeys {
					if _, ok := m[k]; !ok {
						t.Errorf("expected key %q in result, got keys %v", k, keysOf(m))
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// handleStringChar — escape sequences within quoted strings
// ---------------------------------------------------------------------------

func TestFixJSON_EscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkKey string
		wantVal  string
	}{
		{
			name:     "escaped double quote inside string",
			input:    `{"msg": "say \"hello\""}`,
			checkKey: "msg",
			wantVal:  `say "hello"`,
		},
		{
			name:     "escaped newline",
			input:    `{"text": "line1\nline2"}`,
			checkKey: "text",
			wantVal:  "line1\nline2",
		},
		{
			name:     "multiple consecutive escaped backslashes",
			input:    `{"a": "\\\\"}`,
			checkKey: "a",
			wantVal:  `\\`,
		},
		{
			name:     "escaped tab character",
			input:    `{"t": "col1\tcol2"}`,
			checkKey: "t",
			wantVal:  "col1\tcol2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			m := mustUnmarshalObject(t, got)
			val, ok := m[tt.checkKey]
			if !ok {
				t.Fatalf("missing key %q in result", tt.checkKey)
			}
			if s, _ := val.(string); s != tt.wantVal {
				t.Errorf("key %q: got %q, want %q", tt.checkKey, s, tt.wantVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// handleNonStringChar — backtick-quoted strings
// ---------------------------------------------------------------------------

func TestFixJSON_BacktickQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkKey string
		wantVal  string
	}{
		{
			name:     "backtick keys and values",
			input:    "{`name`: `Alice`}",
			checkKey: "name",
			wantVal:  "Alice",
		},
		{
			name:     "mixed double and backtick quotes",
			input:    "{\"name\": `Bob`}",
			checkKey: "name",
			wantVal:  "Bob",
		},
		{
			name:     "backtick key with double-quoted value",
			input:    "{`age`: 42}",
			checkKey: "age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			m := mustUnmarshalObject(t, got)
			val, ok := m[tt.checkKey]
			if !ok {
				t.Fatalf("missing key %q in result; keys: %v", tt.checkKey, keysOf(m))
			}
			if tt.wantVal != "" {
				if s, _ := val.(string); s != tt.wantVal {
					t.Errorf("key %q: got %q, want %q", tt.checkKey, s, tt.wantVal)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Mixed quote styles: single, double, and backtick in one input
// ---------------------------------------------------------------------------

func TestFixJSON_MixedQuoteStyles(t *testing.T) {
	input := "{\"first\": `Bob`, 'age': 30}"

	got, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON returned error: %v", err)
	}
	mustBeValidJSON(t, got)

	m := mustUnmarshalObject(t, got)
	if s, _ := m["first"].(string); s != "Bob" {
		t.Errorf("expected first=Bob, got %q", s)
	}
	if v, _ := m["age"].(float64); v != 30 {
		t.Errorf("expected age=30, got %v", v)
	}
}

// ---------------------------------------------------------------------------
// Block comments
// ---------------------------------------------------------------------------

func TestFixJSON_BlockComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkKey string
		wantVal  string
	}{
		{
			name:     "inline block comment between key and value",
			input:    `{"name": /* a comment */ "Alice"}`,
			checkKey: "name",
			wantVal:  "Alice",
		},
		{
			name: "multiline block comment",
			input: `{
				"title": /* this is
				a multi-line
				comment */ "hello"
			}`,
			checkKey: "title",
			wantVal:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			m := mustUnmarshalObject(t, got)
			if s, _ := m[tt.checkKey].(string); s != tt.wantVal {
				t.Errorf("key %q: got %q, want %q", tt.checkKey, s, tt.wantVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Line comments
// ---------------------------------------------------------------------------

func TestFixJSON_LineComments(t *testing.T) {
	input := `{
		"a": 1, // this is a comment
		"b": 2
	}`

	got, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON returned error: %v", err)
	}
	mustBeValidJSON(t, got)

	m := mustUnmarshalObject(t, got)
	if v, _ := m["a"].(float64); v != 1 {
		t.Errorf("expected a=1, got %v", v)
	}
	if v, _ := m["b"].(float64); v != 2 {
		t.Errorf("expected b=2, got %v", v)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestFixJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty object",
			input: `{}`,
		},
		{
			name:  "empty array",
			input: `[]`,
		},
		{
			name:  "already valid simple object",
			input: `{"a": 1}`,
		},
		{
			name: "deeply nested valid JSON",
			input: `{
				"l1": {
					"l2": {
						"l3": {
							"l4": {
								"l5": "deep"
							}
						}
					}
				}
			}`,
		},
		{
			name:  "null value",
			input: `{"a": null}`,
		},
		{
			name:  "boolean values",
			input: `{"t": true, "f": false}`,
		},
		{
			name:  "numeric values",
			input: `{"int": 42, "neg": -7, "float": 3.14}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Combined: multiple fix types in a single input
// ---------------------------------------------------------------------------

func TestFixJSON_CombinedFixes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name:     "unquoted keys + single quotes + trailing comma",
			input:    `{name: 'Alice', age: 30,}`,
			wantKeys: []string{"name", "age"},
		},
		{
			name:     "comments + unquoted keys + unclosed",
			input:    `{name: "Bob", // a comment` + "\n" + `age: 25`,
			wantKeys: []string{"name", "age"},
		},
		{
			name:     "backtick + trailing comma + unclosed array",
			input:    "{`items`: [1, 2, 3,",
			wantKeys: []string{"items"},
		},
		{
			name: "block comment + single quotes + nested unclosed",
			input: `{
				'user': /* metadata */ {
					'name': 'Carol'
			`,
			wantKeys: []string{"user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			m := mustUnmarshalObject(t, got)
			for _, k := range tt.wantKeys {
				if _, ok := m[k]; !ok {
					t.Errorf("expected key %q in result, got keys %v", k, keysOf(m))
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Idempotency: fixing already-valid JSON should not corrupt it
// ---------------------------------------------------------------------------

func TestFixJSON_Idempotency(t *testing.T) {
	inputs := []string{
		`[1, 2, 3]`,
		`{"escaped": "line1\nline2\ttab"}`,
		`{"a": 1}`,
	}

	for _, input := range inputs {
		t.Run("", func(t *testing.T) {
			first, err := FixJSON(input)
			if err != nil {
				t.Fatalf("first FixJSON call failed: %v", err)
			}

			second, err := FixJSON(first)
			if err != nil {
				t.Fatalf("second FixJSON call failed: %v", err)
			}

			// Both passes should produce valid JSON.
			mustBeValidJSON(t, first)
			mustBeValidJSON(t, second)

			// Unmarshal both and compare semantically rather than string-equal,
			// because whitespace may differ.
			var v1, v2 interface{}
			json.Unmarshal([]byte(first), &v1)
			json.Unmarshal([]byte(second), &v2)

			b1, _ := json.Marshal(v1)
			b2, _ := json.Marshal(v2)
			if string(b1) != string(b2) {
				t.Errorf("idempotency violated:\n  first:  %s\n  second: %s", first, second)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Regression: strings containing structural characters
// ---------------------------------------------------------------------------

func TestFixJSON_StructuralCharsInStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkKey string
		wantVal  string
	}{
		{
			name:     "braces inside string value",
			input:    `{"template": "Hello {world}"}`,
			checkKey: "template",
			wantVal:  "Hello {world}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			m := mustUnmarshalObject(t, got)
			if s, _ := m[tt.checkKey].(string); s != tt.wantVal {
				t.Errorf("key %q: got %q, want %q", tt.checkKey, s, tt.wantVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Trailing comma removal in various positions
// ---------------------------------------------------------------------------

func TestFixJSON_TrailingCommas(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "trailing comma in object",
			input: `{"a": 1, "b": 2,}`,
		},
		{
			name:  "trailing comma in array",
			input: `[1, 2, 3,]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			// The result must not contain ",}" or ",]" patterns.
			if strings.Contains(got, ",}") || strings.Contains(got, ",]") {
				t.Errorf("trailing comma not removed: %s", got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unquoted keys and values
// ---------------------------------------------------------------------------

func TestFixJSON_UnquotedKeysAndValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name:     "unquoted keys with quoted values",
			input:    `{name: "Alice", city: "NYC"}`,
			wantKeys: []string{"name", "city"},
		},
		{
			name:     "unquoted keys with numeric values",
			input:    `{count: 42, score: 99}`,
			wantKeys: []string{"count", "score"},
		},
		{
			name:     "unquoted keys with boolean values",
			input:    `{active: true, deleted: false}`,
			wantKeys: []string{"active", "deleted"},
		},
		{
			name:     "unquoted keys with null value",
			input:    `{value: null}`,
			wantKeys: []string{"value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			m := mustUnmarshalObject(t, got)
			for _, k := range tt.wantKeys {
				if _, ok := m[k]; !ok {
					t.Errorf("expected key %q in result, got keys %v", k, keysOf(m))
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
