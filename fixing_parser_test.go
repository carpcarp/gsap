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
//
// Coverage target: closeUnclosedStructures (9.1% -> high).
// Each test feeds FixJSON an input missing one or more closing brackets.
// The parser's closeUnclosedStructures method must auto-close them in the
// correct (reverse) order, optionally removing trailing commas and quoting
// unquoted values along the way.
// ---------------------------------------------------------------------------

func TestFixJSON_CloseUnclosedStructures(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKeys  []string // expected top-level keys (nil to skip)
		wantArray bool     // true if the result should be an array
		wantLen   int      // expected array length (-1 to skip)
	}{
		{
			name:     "unclosed object with quoted value",
			input:    `{"name": "Alice"`,
			wantKeys: []string{"name"},
		},
		{
			name:      "unclosed array of strings",
			input:     `["a", "b"`,
			wantArray: true,
			wantLen:   2,
		},
		{
			name:     "nested unclosed array of strings inside object",
			input:    `{"items": ["a", "b"`,
			wantKeys: []string{"items"},
		},
		{
			name:     "unclosed object with trailing comma and quoted values",
			input:    `{"a": "x", "b": "y",`,
			wantKeys: []string{"a", "b"},
		},
		{
			name:     "multiple nested unclosed with string leaf",
			input:    `{"a": {"b": ["hello"`,
			wantKeys: []string{"a"},
		},
		{
			name:     "deeply nested all unclosed with string values",
			input:    `{"l1": {"l2": {"l3": ["deep"`,
			wantKeys: []string{"l1"},
		},
		{
			name:      "unclosed array of strings with trailing comma",
			input:     `["x", "y", "z",`,
			wantArray: true,
			wantLen:   3,
		},
		{
			name:      "unclosed top-level array of numbers",
			input:     `[1, 2, 3`,
			wantArray: true,
			wantLen:   3,
		},
		{
			name:     "unclosed object with multiple string pairs",
			input:    `{"first": "Alice", "last": "Smith"`,
			wantKeys: []string{"first", "last"},
		},
		{
			name:     "unclosed object after comma with no next pair",
			input:    `{"a": "one",`,
			wantKeys: []string{"a"},
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
//
// Coverage target: handleStringChar (57% -> high).
// Missing paths: stringEscaped branch (backslash sets the flag, next char
// clears it), and the quote-char branch (stringQuoteChar terminates the
// string and writes a double-quote regardless of original quote style).
// ---------------------------------------------------------------------------

func TestFixJSON_EscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkKey string
		wantVal  string
	}{
		{
			name:     "escaped double quote preserves inner quotes",
			input:    `{"msg": "say \"hello\""}`,
			checkKey: "msg",
			wantVal:  `say "hello"`,
		},
		{
			name:     "escaped backslash before closing quote",
			input:    `{"a": "test\\"}`,
			checkKey: "a",
			wantVal:  `test\`,
		},
		{
			name:     "escaped newline character",
			input:    `{"text": "line1\nline2"}`,
			checkKey: "text",
			wantVal:  "line1\nline2",
		},
		{
			name:     "double escaped backslash",
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
		{
			name:     "escaped quote at start of value",
			input:    `{"q": "\"quoted\""}`,
			checkKey: "q",
			wantVal:  `"quoted"`,
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
// handleStringChar — stringEscaped flag transitions
//
// Verifies the escaped-character state machine toggles correctly:
// backslash sets the flag, the very next character clears it, and
// subsequent characters are processed normally.
// ---------------------------------------------------------------------------

func TestFixJSON_EscapeStateMachine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkKey string
		wantVal  string
	}{
		{
			name:     "escape then normal char then closing quote",
			input:    `{"k": "a\nb"}`,
			checkKey: "k",
			wantVal:  "a\nb",
		},
		{
			name:     "consecutive escapes cancel out",
			input:    `{"k": "a\\b"}`,
			checkKey: "k",
			wantVal:  "a\\b",
		},
		{
			name:     "escape at very end of string value",
			input:    `{"k": "end\\"}`,
			checkKey: "k",
			wantVal:  `end\`,
		},
		{
			name:     "escaped quote does not end string",
			input:    `{"k": "has\"quote"}`,
			checkKey: "k",
			wantVal:  `has"quote`,
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
// handleNonStringChar — backtick-quoted strings converted to double quotes
//
// Coverage target: the '`' case in handleNonStringChar (under-tested).
// Backtick strings should enter string mode with stringQuoteChar = '`'
// and terminate at the closing backtick, emitting double quotes.
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
			name:     "backtick value with double-quoted key",
			input:    "{\"name\": `Bob`}",
			checkKey: "name",
			wantVal:  "Bob",
		},
		{
			name:     "backtick key with numeric value",
			input:    "{`count`: 7}",
			checkKey: "count",
		},
		{
			name:     "backtick with spaces in value",
			input:    "{`greeting`: `hello world`}",
			checkKey: "greeting",
			wantVal:  "hello world",
		},
		{
			name:     "multiple backtick pairs",
			input:    "{`a`: `x`, `b`: `y`}",
			checkKey: "a",
			wantVal:  "x",
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
// All three quote styles mixed in a single input
// ---------------------------------------------------------------------------

func TestFixJSON_MixedQuoteStyles(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name:     "double + backtick + single in one object",
			input:    "{\"first\": `Bob`, 'last': 'Smith'}",
			wantKeys: []string{"first", "last"},
		},
		{
			name:     "single-quoted keys with backtick values",
			input:    "{'a': `x`, 'b': `y`}",
			wantKeys: []string{"a", "b"},
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
					t.Errorf("expected key %q, got keys %v", k, keysOf(m))
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Block comments (/* ... */)
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
		{
			name:     "block comment before first key",
			input:    `{/* intro */ "x": "y"}`,
			checkKey: "x",
			wantVal:  "y",
		},
		{
			name:     "block comment after last value",
			input:    `{"k": "v" /* trailing */}`,
			checkKey: "k",
			wantVal:  "v",
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
// Line comments (// ...)
// ---------------------------------------------------------------------------

func TestFixJSON_LineComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name: "line comment after value",
			input: `{
				"a": "one", // first
				"b": "two"
			}`,
			wantKeys: []string{"a", "b"},
		},
		{
			name: "line comment on its own line",
			input: `{
				// this is a header comment
				"x": "y"
			}`,
			wantKeys: []string{"x"},
		},
		{
			name: "multiple line comments",
			input: `{
				"p": "q", // comment 1
				// comment 2
				"r": "s"  // comment 3
			}`,
			wantKeys: []string{"p", "r"},
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
					t.Errorf("expected key %q, got keys %v", k, keysOf(m))
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Edge cases: valid JSON and boundary conditions
// ---------------------------------------------------------------------------

func TestFixJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty object", `{}`},
		{"empty array", `[]`},
		{"single key-value pair", `{"a": "b"}`},
		{"deeply nested objects with string values", `{
			"l1": {
				"l2": {
					"l3": {
						"l4": {
							"l5": "deep"
						}
					}
				}
			}
		}`},
		{"boolean true", `{"t": true}`},
		{"boolean false", `{"f": false}`},
		{"null value", `{"n": null}`},
		{"integer value", `{"x": 42}`},
		{"negative integer", `{"x": -7}`},
		{"float value", `{"x": 3.14}`},
		{"string with spaces", `{"msg": "hello world"}`},
		{"array of strings", `["a", "b", "c"]`},
		{"top-level number array", `[1, 2, 3]`},
		{"nested arrays of strings", `{"m": [["a", "b"], ["c"]]}`},
		{"empty string value", `{"e": ""}`},
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
// Combined: multiple fix types applied in a single input
// ---------------------------------------------------------------------------

func TestFixJSON_CombinedFixes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name:     "unquoted keys + single-quoted values + trailing comma",
			input:    `{name: 'Alice', city: 'NYC',}`,
			wantKeys: []string{"name", "city"},
		},
		{
			name:     "line comment + unquoted keys + unclosed",
			input:    `{name: "Bob", // inline` + "\n" + `age: 25`,
			wantKeys: []string{"name", "age"},
		},
		{
			name:     "backtick + trailing comma + unclosed array of strings",
			input:    "{`items`: [\"a\", \"b\",",
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
		{
			name:     "unquoted key + backtick value + line comment",
			input:    "{status: `active` // note\n}",
			wantKeys: []string{"status"},
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
// Idempotency: running FixJSON on its own output should not corrupt data
// ---------------------------------------------------------------------------

func TestFixJSON_Idempotency(t *testing.T) {
	inputs := []string{
		`{"name": "Alice"}`,
		`[1, 2, 3]`,
		`{"escaped": "line1\nline2\ttab"}`,
		`["hello", "world"]`,
		`{"flag": true}`,
		`{"empty": ""}`,
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

			mustBeValidJSON(t, first)
			mustBeValidJSON(t, second)

			// Compare semantically (whitespace may differ between passes).
			var v1, v2 interface{}
			_ = json.Unmarshal([]byte(first), &v1)
			_ = json.Unmarshal([]byte(second), &v2)

			b1, _ := json.Marshal(v1)
			b2, _ := json.Marshal(v2)
			if string(b1) != string(b2) {
				t.Errorf("idempotency violated:\n  first:  %s\n  second: %s", first, second)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Structural characters inside quoted strings must be preserved
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
		{
			name:     "backslash-n in string",
			input:    `{"nl": "a\nb"}`,
			checkKey: "nl",
			wantVal:  "a\nb",
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
// Trailing comma removal
// ---------------------------------------------------------------------------

func TestFixJSON_TrailingCommas(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "trailing comma in flat object with string values",
			input: `{"a": "x", "b": "y",}`,
		},
		{
			name:  "trailing comma in array of strings",
			input: `["a", "b", "c",]`,
		},
		{
			name:  "trailing comma in array of numbers",
			input: `[1, 2, 3,]`,
		},
		{
			name:  "trailing comma after boolean",
			input: `{"flag": true,}`,
		},
		{
			name:  "trailing comma after null",
			input: `{"val": null,}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixJSON(tt.input)
			if err != nil {
				t.Fatalf("FixJSON returned error: %v", err)
			}
			mustBeValidJSON(t, got)

			if strings.Contains(got, ",}") || strings.Contains(got, ",]") {
				t.Errorf("trailing comma not fully removed: %s", got)
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
			name:     "unquoted keys with quoted string values",
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
		{
			name:     "unquoted keys with unquoted string values",
			input:    `{name: Bob, status: active}`,
			wantKeys: []string{"name", "status"},
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
// handleStringChar — single-quoted strings end on single quote, not double
// ---------------------------------------------------------------------------

func TestFixJSON_SingleQuoteStringBoundary(t *testing.T) {
	// Verify that a single-quoted string terminates at the matching single
	// quote and produces valid double-quoted JSON output.
	input := `{'greeting': 'hello world'}`

	got, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON returned error: %v", err)
	}
	mustBeValidJSON(t, got)

	m := mustUnmarshalObject(t, got)
	s, _ := m["greeting"].(string)
	if s != "hello world" {
		t.Errorf("expected 'hello world', got %q", s)
	}
}

// ---------------------------------------------------------------------------
// handleStringChar — backtick strings end on backtick, not double quote
// ---------------------------------------------------------------------------

func TestFixJSON_BacktickStringBoundary(t *testing.T) {
	// Verify that a backtick-quoted string terminates at the matching
	// backtick and produces valid double-quoted JSON output.
	input := "{`greeting`: `hello world`}"

	got, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON returned error: %v", err)
	}
	mustBeValidJSON(t, got)

	m := mustUnmarshalObject(t, got)
	s, _ := m["greeting"].(string)
	if s != "hello world" {
		t.Errorf("expected 'hello world', got %q", s)
	}
}

// ---------------------------------------------------------------------------
// FixJSON always returns a non-error result
// ---------------------------------------------------------------------------

func TestFixJSON_NeverReturnsError(t *testing.T) {
	// The parser is designed to always produce output, even for degenerate
	// inputs. Verify it never returns an error.
	inputs := []string{
		``,
		`{}`,
		`[]`,
		`{`,
		`[`,
		`"hello"`,
		`42`,
		`true`,
		`null`,
		`{{{`,
	}

	for _, input := range inputs {
		t.Run("", func(t *testing.T) {
			_, err := FixJSON(input)
			if err != nil {
				t.Errorf("FixJSON(%q) returned unexpected error: %v", input, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Truncation scenarios that closeUnclosedStructures can handle
//
// The parser closes structures when truncation happens outside a string
// (i.e., the last string was already terminated). Mid-string truncation
// is a known limitation.
// ---------------------------------------------------------------------------

func TestFixJSON_TruncatedAfterCompleteValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name:     "truncated after complete key-value pair",
			input:    `{"name": "Alice"`,
			wantKeys: []string{"name"},
		},
		{
			name:     "truncated after two complete pairs",
			input:    `{"name": "Alice", "age": 30`,
			wantKeys: []string{"name"},
		},
		{
			name:     "truncated after comma with no next value",
			input:    `{"name": "Alice",`,
			wantKeys: []string{"name"},
		},
		{
			name:     "truncated nested with complete inner string",
			input:    `{"user": {"name": "Bob"`,
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
// Value preservation: verify specific values survive the fixing process
// ---------------------------------------------------------------------------

func TestFixJSON_ValuePreservation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		key      string
		wantStr  string
		wantNum  float64
		wantBool *bool
		isNum    bool
		isBool   bool
	}{
		{
			name:    "quoted string preserved exactly",
			input:   `{"msg": "Hello World"}`,
			key:     "msg",
			wantStr: "Hello World",
		},
		{
			name:    "single-quoted string value preserved",
			input:   `{'msg': 'testing'}`,
			key:     "msg",
			wantStr: "testing",
		},
		{
			name:    "backtick string value preserved",
			input:   "{`msg`: `works`}",
			key:     "msg",
			wantStr: "works",
		},
		{
			name:    "integer value preserved",
			input:   `{"n": 42}`,
			key:     "n",
			wantNum: 42,
			isNum:   true,
		},
		{
			name:    "negative number preserved",
			input:   `{"n": -10}`,
			key:     "n",
			wantNum: -10,
			isNum:   true,
		},
		{
			name:     "boolean true preserved",
			input:    `{"ok": true}`,
			key:      "ok",
			wantBool: boolPtr(true),
			isBool:   true,
		},
		{
			name:     "boolean false preserved",
			input:    `{"ok": false}`,
			key:      "ok",
			wantBool: boolPtr(false),
			isBool:   true,
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
			val, ok := m[tt.key]
			if !ok {
				t.Fatalf("missing key %q", tt.key)
			}

			switch {
			case tt.isNum:
				if n, _ := val.(float64); n != tt.wantNum {
					t.Errorf("key %q: got %v, want %v", tt.key, n, tt.wantNum)
				}
			case tt.isBool:
				if b, _ := val.(bool); b != *tt.wantBool {
					t.Errorf("key %q: got %v, want %v", tt.key, b, *tt.wantBool)
				}
			default:
				if s, _ := val.(string); s != tt.wantStr {
					t.Errorf("key %q: got %q, want %q", tt.key, s, tt.wantStr)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func boolPtr(b bool) *bool {
	return &b
}
