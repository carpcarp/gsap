package sap

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// coerceValueToString
// ---------------------------------------------------------------------------

func TestCoerceValueToString(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "string passthrough",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "float64 whole number",
			input: float64(42),
			want:  "42",
		},
		{
			name:  "float64 negative whole number",
			input: float64(-7),
			want:  "-7",
		},
		{
			name:  "float64 zero",
			input: float64(0),
			want:  "0",
		},
		{
			name:  "float64 fractional",
			input: 3.14,
			want:  "3.140000",
		},
		{
			name:  "bool true",
			input: true,
			want:  "true",
		},
		{
			name:  "bool false",
			input: false,
			want:  "false",
		},
		{
			name:    "unsupported type slice",
			input:   []int{1, 2, 3},
			wantErr: true,
		},
		{
			name:    "unsupported type map",
			input:   map[string]int{"a": 1},
			wantErr: true,
		},
		{
			name:    "unsupported type nil",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := coerceValueToString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// levenshteinDistance
// ---------------------------------------------------------------------------

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want int
	}{
		{
			name: "identical strings",
			s1:   "kitten",
			s2:   "kitten",
			want: 0,
		},
		{
			name: "both empty",
			s1:   "",
			s2:   "",
			want: 0,
		},
		{
			name: "first empty",
			s1:   "",
			s2:   "abc",
			want: 3,
		},
		{
			name: "second empty",
			s1:   "abc",
			s2:   "",
			want: 3,
		},
		{
			name: "single insertion",
			s1:   "cat",
			s2:   "cats",
			want: 1,
		},
		{
			name: "single deletion",
			s1:   "cats",
			s2:   "cat",
			want: 1,
		},
		{
			name: "single substitution",
			s1:   "cat",
			s2:   "car",
			want: 1,
		},
		{
			name: "classic kitten to sitting",
			s1:   "kitten",
			s2:   "sitting",
			want: 3,
		},
		{
			name: "completely different",
			s1:   "abc",
			s2:   "xyz",
			want: 3,
		},
		{
			name: "single character vs empty",
			s1:   "a",
			s2:   "",
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := levenshteinDistance(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// normalizeString
// ---------------------------------------------------------------------------

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain ASCII lowercase",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "plain ASCII mixed case",
			input: "Hello World",
			want:  "hello world",
		},
		{
			name:  "accented e in etude",
			input: "étude",
			want:  "etude",
		},
		{
			name:  "naive with diaeresis",
			input: "naïve",
			want:  "naive",
		},
		{
			name:  "ligature ae",
			input: "Cæsar",
			want:  "casar",
		},
		{
			name:  "ligature oe",
			input: "cœur",
			want:  "cour",
		},
		{
			name:  "eszett mapped to single s",
			input: "Straße",
			want:  "strase",
		},
		{
			name:  "cedilla",
			input: "façade",
			want:  "facade",
		},
		{
			name:  "tilde n",
			input: "jalapeño",
			want:  "jalapeno",
		},
		{
			name:  "punctuation stripped",
			input: "hello, world! #2024",
			want:  "hello world 2024",
		},
		{
			name:  "underscore preserved",
			input: "snake_case",
			want:  "snake_case",
		},
		{
			name:  "digits preserved",
			input: "abc123",
			want:  "abc123",
		},
		{
			name:  "multiple accented characters",
			input: "Ángélíqué",
			want:  "angelique",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeString(tt.input)
			if got != tt.want {
				t.Errorf("normalizeString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// stringDistance
// ---------------------------------------------------------------------------

func TestStringDistance(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want int
	}{
		{
			name: "normalization makes equal",
			s1:   "Étude",
			s2:   "etude",
			want: 0,
		},
		{
			name: "accented vs plain same word",
			s1:   "café",
			s2:   "cafe",
			want: 0,
		},
		{
			name: "case difference normalized away",
			s1:   "CANCELED",
			s2:   "canceled",
			want: 0,
		},
		{
			name: "close after normalization",
			s1:   "Cancelled",
			s2:   "Canceled",
			want: 1,
		},
		{
			name: "different words",
			s1:   "apple",
			s2:   "orange",
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringDistance(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("stringDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// fuzzyMatchEnum
// ---------------------------------------------------------------------------

func TestFuzzyMatchEnum(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		enumValues []string
		want       string
	}{
		{
			name:       "exact match in list",
			input:      "Active",
			enumValues: []string{"Active", "Canceled", "Pending"},
			want:       "Active",
		},
		{
			name:       "close typo match",
			input:      "Cancelld",
			enumValues: []string{"Canceled", "Active"},
			want:       "Canceled",
		},
		{
			name:       "no match within threshold",
			input:      "xylophone",
			enumValues: []string{"Canceled", "Active"},
			want:       "",
		},
		{
			name:       "empty enum values",
			input:      "anything",
			enumValues: []string{},
			want:       "",
		},
		{
			name:       "nil enum values",
			input:      "anything",
			enumValues: nil,
			want:       "",
		},
		{
			name:       "picks best match among multiple",
			input:      "Pnding",
			enumValues: []string{"Active", "Canceled", "Pending"},
			want:       "Pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzyMatchEnum(tt.input, tt.enumValues)
			if got != tt.want {
				t.Errorf("fuzzyMatchEnum(%q, %v) = %q, want %q", tt.input, tt.enumValues, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CoerceToEnum
// ---------------------------------------------------------------------------

func TestCoerceToEnum(t *testing.T) {
	// getEnumValues returns an empty slice, so CoerceToEnum always falls
	// through to returning the stringified input. We verify it doesn't
	// panic and returns the expected stringified value.
	enumType := reflect.TypeOf("")

	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:  "string input",
			input: "Active",
			want:  "Active",
		},
		{
			name:  "float64 whole number",
			input: float64(42),
			want:  "42",
		},
		{
			name:  "float64 fractional",
			input: 3.14,
			want:  "3.140000",
		},
		{
			name:  "bool true",
			input: true,
			want:  "true",
		},
		{
			name:  "bool false",
			input: false,
			want:  "false",
		},
		{
			name:    "unsupported type returns error",
			input:   []int{1, 2},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := &Score{flags: make(map[string]int)}
			got, err := CoerceToEnum(tt.input, enumType, score)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotStr, ok := got.(string)
			if !ok {
				t.Fatalf("expected string result, got %T", got)
			}
			if gotStr != tt.want {
				t.Errorf("got %q, want %q", gotStr, tt.want)
			}

			// With empty enum values, no flags should be set.
			if score.Total() != 0 {
				t.Errorf("expected score total 0 (no enum match path taken), got %d", score.Total())
			}
		})
	}
}

// TestCoerceToEnumScoreNotMutated verifies that passing a score with
// existing flags doesn't cause unexpected side effects.
func TestCoerceToEnumScoreNotMutated(t *testing.T) {
	score := &Score{flags: make(map[string]int)}
	score.AddFlag("PreExisting", 5)

	enumType := reflect.TypeOf("")
	_, err := CoerceToEnum("test", enumType, score)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score.Total() != 5 {
		t.Errorf("expected score total to remain 5, got %d", score.Total())
	}
}
