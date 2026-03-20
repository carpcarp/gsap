package sap

import (
	"math"
	"reflect"
	"testing"
)

func newScore() *Score {
	return &Score{flags: make(map[string]int)}
}

// ---------------------------------------------------------------------------
// coerceToString
// ---------------------------------------------------------------------------

func TestCoerceToString(t *testing.T) {
	c := NewTypeCoercer()

	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"string passthrough", "hello", "hello"},
		{"empty string", "", ""},
		{"whole float64", float64(42), "42"},
		{"negative whole float64", float64(-7), "-7"},
		{"fractional float64", 3.14, "3.14"},
		{"zero float64", float64(0), "0"},
		{"large whole float64", float64(1000000), "1000000"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"int via default", 99, "99"},           // goes through fmt.Sprintf
		{"nil slice via default", []int(nil), "[]"}, // exercises default branch
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.coerceToString(tt.input, newScore())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("coerceToString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// coerceToUint
// ---------------------------------------------------------------------------

func TestCoerceToUint(t *testing.T) {
	c := NewTypeCoercer()

	tests := []struct {
		name       string
		input      interface{}
		targetType reflect.Type
		want       uint64
	}{
		{"float64 to uint64", float64(42), reflect.TypeOf(uint64(0)), 42},
		{"float64 zero", float64(0), reflect.TypeOf(uint64(0)), 0},
		{"string number to uint64", "100", reflect.TypeOf(uint64(0)), 100},
		{"bool true to uint64", true, reflect.TypeOf(uint64(0)), 1},
		{"bool false to uint64", false, reflect.TypeOf(uint64(0)), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.coerceToUint(tt.input, tt.targetType, newScore())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// The result is returned via reflect.Convert, so we need to compare through uint64.
			gotUint := reflect.ValueOf(got).Convert(reflect.TypeOf(uint64(0))).Uint()
			if gotUint != tt.want {
				t.Errorf("coerceToUint(%v) = %d, want %d", tt.input, gotUint, tt.want)
			}
		})
	}

	// Smaller unsigned types.
	t.Run("uint8 target", func(t *testing.T) {
		got, err := c.coerceToUint(float64(255), reflect.TypeOf(uint8(0)), newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.(uint8) != 255 {
			t.Errorf("expected uint8(255), got %v", got)
		}
	})

	t.Run("uint32 target", func(t *testing.T) {
		got, err := c.coerceToUint(float64(1000), reflect.TypeOf(uint32(0)), newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.(uint32) != 1000 {
			t.Errorf("expected uint32(1000), got %v", got)
		}
	})

	// Error path: non-numeric input.
	t.Run("error on struct input", func(t *testing.T) {
		_, err := c.coerceToUint(struct{}{}, reflect.TypeOf(uint64(0)), newScore())
		if err == nil {
			t.Fatal("expected error for non-numeric input, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// coerceToArray
// ---------------------------------------------------------------------------

func TestCoerceToArray(t *testing.T) {
	c := NewTypeCoercer()

	t.Run("slice of interface to [3]int", func(t *testing.T) {
		input := []interface{}{float64(1), float64(2), float64(3)}
		target := reflect.TypeOf([3]int{})
		got, err := c.coerceToArray(input, target, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr := got.([3]int)
		if arr != [3]int{1, 2, 3} {
			t.Errorf("got %v, want [1 2 3]", arr)
		}
	})

	t.Run("fewer items than array length pads with zero", func(t *testing.T) {
		input := []interface{}{float64(10)}
		target := reflect.TypeOf([3]int{})
		got, err := c.coerceToArray(input, target, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr := got.([3]int)
		if arr != [3]int{10, 0, 0} {
			t.Errorf("got %v, want [10 0 0]", arr)
		}
	})

	t.Run("more items than array length truncates", func(t *testing.T) {
		input := []interface{}{float64(1), float64(2), float64(3), float64(4)}
		target := reflect.TypeOf([2]int{})
		got, err := c.coerceToArray(input, target, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr := got.([2]int)
		if arr != [2]int{1, 2} {
			t.Errorf("got %v, want [1 2]", arr)
		}
	})

	t.Run("single non-slice value wraps into array", func(t *testing.T) {
		input := float64(42)
		target := reflect.TypeOf([3]int{})
		got, err := c.coerceToArray(input, target, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr := got.([3]int)
		if arr[0] != 42 {
			t.Errorf("first element = %d, want 42", arr[0])
		}
	})

	t.Run("string array", func(t *testing.T) {
		input := []interface{}{"a", "b"}
		target := reflect.TypeOf([2]string{})
		got, err := c.coerceToArray(input, target, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr := got.([2]string)
		if arr != [2]string{"a", "b"} {
			t.Errorf("got %v, want [a b]", arr)
		}
	})

	t.Run("error propagation from element coercion", func(t *testing.T) {
		input := []interface{}{struct{}{}}
		target := reflect.TypeOf([1]int{})
		_, err := c.coerceToArray(input, target, newScore())
		if err == nil {
			t.Fatal("expected error when element cannot be coerced, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// coerceToMap
// ---------------------------------------------------------------------------

func TestCoerceToMap(t *testing.T) {
	c := NewTypeCoercer()

	t.Run("map[string]interface{} to map[string]string", func(t *testing.T) {
		input := map[string]interface{}{"key": "value", "foo": "bar"}
		target := reflect.TypeOf(map[string]string{})
		got, err := c.coerceToMap(input, target, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m := got.(map[string]string)
		if m["key"] != "value" || m["foo"] != "bar" {
			t.Errorf("got %v, want map[key:value foo:bar]", m)
		}
	})

	t.Run("map with float64 values coerced to int", func(t *testing.T) {
		input := map[string]interface{}{"count": float64(5), "total": float64(10)}
		target := reflect.TypeOf(map[string]int{})
		got, err := c.coerceToMap(input, target, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m := got.(map[string]int)
		if m["count"] != 5 || m["total"] != 10 {
			t.Errorf("got %v, want map[count:5 total:10]", m)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		input := map[string]interface{}{}
		target := reflect.TypeOf(map[string]string{})
		got, err := c.coerceToMap(input, target, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m := got.(map[string]string)
		if len(m) != 0 {
			t.Errorf("expected empty map, got %v", m)
		}
	})

	t.Run("error on non-map input", func(t *testing.T) {
		_, err := c.coerceToMap("not a map", reflect.TypeOf(map[string]string{}), newScore())
		if err == nil {
			t.Fatal("expected error for non-map input, got nil")
		}
	})

	t.Run("error on int input", func(t *testing.T) {
		_, err := c.coerceToMap(42, reflect.TypeOf(map[string]string{}), newScore())
		if err == nil {
			t.Fatal("expected error for int input, got nil")
		}
	})

	t.Run("error on slice input", func(t *testing.T) {
		_, err := c.coerceToMap([]interface{}{1, 2}, reflect.TypeOf(map[string]string{}), newScore())
		if err == nil {
			t.Fatal("expected error for slice input, got nil")
		}
	})

	t.Run("value coercion error propagates", func(t *testing.T) {
		// struct{}{} cannot be coerced to int
		input := map[string]interface{}{"bad": struct{}{}}
		target := reflect.TypeOf(map[string]int{})
		_, err := c.coerceToMap(input, target, newScore())
		if err == nil {
			t.Fatal("expected error when value cannot be coerced, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// coerceToBool — under-tested paths
// ---------------------------------------------------------------------------

func TestCoerceToBool(t *testing.T) {
	c := NewTypeCoercer()

	t.Run("bool passthrough true", func(t *testing.T) {
		got, err := c.coerceToBool(true, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != true {
			t.Errorf("got %v, want true", got)
		}
	})

	t.Run("bool passthrough false", func(t *testing.T) {
		got, err := c.coerceToBool(false, newScore())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != false {
			t.Errorf("got %v, want false", got)
		}
	})

	// float64 -> bool (0% coverage path)
	float64Tests := []struct {
		name string
		in   float64
		want bool
	}{
		{"nonzero positive", 1.0, true},
		{"nonzero negative", -5.0, true},
		{"nonzero fraction", 0.001, true},
		{"zero", 0.0, false},
	}
	for _, tt := range float64Tests {
		t.Run("float64 "+tt.name, func(t *testing.T) {
			score := newScore()
			got, err := c.coerceToBool(tt.in, score)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("coerceToBool(%v) = %v, want %v", tt.in, got, tt.want)
			}
			if score.Flags()["NumberToBool"] != 1 {
				t.Errorf("expected NumberToBool flag to be set")
			}
		})
	}

	// String variants including markdown-wrapped bools
	stringTests := []struct {
		name    string
		in      string
		want    bool
		wantErr bool
	}{
		{"true", "true", true, false},
		{"false", "false", false, false},
		{"yes", "yes", true, false},
		{"no", "no", false, false},
		{"1", "1", true, false},
		{"0", "0", false, false},
		{"on", "on", true, false},
		{"off", "off", false, false},
		{"y", "y", true, false},
		{"n", "n", false, false},
		{"enabled", "enabled", true, false},
		{"disabled", "disabled", false, false},
		{"active", "active", true, false},
		{"inactive", "inactive", false, false},
		{"uppercase TRUE", "TRUE", true, false},
		{"mixed case Yes", "Yes", true, false},
		{"whitespace padded", "  true  ", true, false},
		{"bold markdown true", "**true**", true, false},
		{"italic markdown false", "*false*", false, false},
		{"code markdown yes", "`yes`", true, false},
		{"invalid string", "maybe", false, true},
		{"invalid string banana", "banana", false, true},
		{"empty string", "", false, true},
	}
	for _, tt := range stringTests {
		t.Run("string "+tt.name, func(t *testing.T) {
			got, err := c.coerceToBool(tt.in, newScore())
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("coerceToBool(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}

	// Unsupported type
	t.Run("error on unsupported type", func(t *testing.T) {
		_, err := c.coerceToBool([]int{1, 2}, newScore())
		if err == nil {
			t.Fatal("expected error for slice input, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// stripMarkdown — all format variants
// ---------------------------------------------------------------------------

func TestStripMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		changed bool
	}{
		{"bold double asterisk", "**hello**", "hello", true},
		{"bold double underscore", "__hello__", "hello", true},
		{"italic single asterisk", "*hello*", "hello", true},
		{"italic single underscore", "_hello_", "hello", true},
		{"strikethrough", "~~hello~~", "hello", true},
		{"inline code", "`hello`", "hello", true},
		{"plain string unchanged", "hello", "hello", false},
		{"number bold", "**42**", "42", true},
		{"number italic underscore", "_30_", "30", true},
		{"empty bold markers", "****", "****", false},     // len <= 4
		{"empty italic markers", "**", "**", false},       // len <= 2
		{"single asterisk", "*", "*", false},              // len <= 2
		{"single backtick", "`", "`", false},              // len <= 2
		{"nested bold and italic", "***text***", "text", true}, // strips ** then *
		{"only whitespace inside bold", "** **", " ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := stripMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("stripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
			if changed != tt.changed {
				t.Errorf("stripMarkdown(%q) changed = %v, want %v", tt.input, changed, tt.changed)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseNumber — edge cases
// ---------------------------------------------------------------------------

func TestParseNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		// Currency symbols
		{"dollar amount", "$100", 100, false},
		{"euro amount", "€50", 50, false},
		{"pound amount", "£75.50", 75.5, false},
		{"dollar with commas", "$1,000,000", 1000000, false},

		// B and T suffixes
		{"B suffix", "3B", 3_000_000_000, false},
		{"T suffix", "1.5T", 1_500_000_000_000, false},
		{"K suffix", "200K", 200_000, false},
		{"M suffix", "2.5M", 2_500_000, false},

		// Fractions
		{"simple fraction", "1/2", 0.5, false},
		{"fraction with spaces", " 3 / 4 ", 0.75, false},
		{"division by zero guard", "5/0", 5, false},     // denom == 0, falls through to trailing-unit regex which captures "5"
		{"invalid fraction numerator", "abc/2", 0, true}, // "abc/2" is not a valid number
		{"invalid fraction denominator", "1/abc", 1, false}, // falls through to trailing-unit regex which captures "1"

		// Negative numbers with suffixes
		{"negative K", "-5K", -5000, false},
		{"negative M", "-2M", -2000000, false},
		{"negative plain", "-42", -42, false},

		// Numbers with unit words
		{"number with percent", "85%", 85, false},
		{"number with unit word", "30 years", 30, false},
		{"number with complex unit", "4 GB", 4, false},
		{"comma number with unit", "1,500 users", 1500, false},

		// Plain numbers
		{"integer", "42", 42, false},
		{"float", "3.14", 3.14, false},
		{"zero", "0", 0, false},
		{"negative float", "-9.81", -9.81, false},

		// Markdown-wrapped
		{"bold number", "**100**", 100, false},
		{"italic number", "_50_", 50, false},

		// Whitespace
		{"leading/trailing spaces", "  42  ", 42, false},

		// Errors
		{"non-numeric string", "hello", 0, true},
		{"empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNumber(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got result %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("parseNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isNullString — all variants and edge cases
// ---------------------------------------------------------------------------

func TestIsNullString(t *testing.T) {
	trueVariants := []string{
		"n/a", "N/A", "na", "NA", "Na",
		"none", "None", "NONE",
		"null", "Null", "NULL",
		"nil", "Nil", "NIL",
		"unknown", "Unknown", "UNKNOWN",
		"tbd", "TBD", "Tbd",
		"undefined", "Undefined", "UNDEFINED",
		"-", "--",
		"  n/a  ",   // with whitespace
		"  null  ",  // with whitespace
		"\tnone\t",  // with tabs
	}

	for _, s := range trueVariants {
		t.Run("true/"+s, func(t *testing.T) {
			if !isNullString(s) {
				t.Errorf("isNullString(%q) = false, want true", s)
			}
		})
	}

	falseVariants := []string{
		"", "hello", "0", "false", "no",
		"actual value", "not null", "---", "n-a",
	}

	for _, s := range falseVariants {
		t.Run("false/"+s, func(t *testing.T) {
			if isNullString(s) {
				t.Errorf("isNullString(%q) = true, want false", s)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// coerceToBool via Coerce (integration-level, checks score flags)
// ---------------------------------------------------------------------------

func TestCoerceToBoolMarkdownStrippedFlag(t *testing.T) {
	c := NewTypeCoercer()
	score := newScore()

	got, err := c.coerceToBool("**true**", score)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Errorf("got %v, want true", got)
	}
	if score.Flags()[FlagMarkdownStripped] != 1 {
		t.Errorf("expected MarkdownStripped flag, got flags: %v", score.Flags())
	}
	if score.Flags()[FlagStringToBool] != 1 {
		t.Errorf("expected StringToBool flag, got flags: %v", score.Flags())
	}
}

// ---------------------------------------------------------------------------
// Coerce (public API) — integration with score
// ---------------------------------------------------------------------------

func TestCoercePublicAPI(t *testing.T) {
	c := NewTypeCoercer()

	t.Run("nil value returns nil", func(t *testing.T) {
		got, score, err := c.Coerce(nil, reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
		if score.Total() != 0 {
			t.Errorf("expected score 0, got %d", score.Total())
		}
	})

	t.Run("string to int returns score", func(t *testing.T) {
		got, score, err := c.Coerce("42", reflect.TypeOf(int(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.(int) != 42 {
			t.Errorf("got %v, want 42", got)
		}
		if score.Total() == 0 {
			t.Error("expected nonzero score for string-to-int coercion")
		}
	})
}
