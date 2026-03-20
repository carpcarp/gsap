package sap

import (
	"encoding/json"
	"reflect"
	"testing"
)

type TestUser struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type TestResume struct {
	Title      string   `json:"title"`
	Experience []string `json:"experience"`
	Active     bool     `json:"active"`
}

func TestParseValidJSON(t *testing.T) {
	input := `{"name": "John", "age": 30, "email": "john@example.com"}`

	user, err := Parse[TestUser](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if user.Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", user.Name)
	}
	if user.Age != 30 {
		t.Errorf("Expected age 30, got %d", user.Age)
	}
}

func TestParseJSONInMarkdown(t *testing.T) {
	input := "Here's the extracted data:\n```json\n{\n  \"name\": \"Alice\",\n  \"age\": 25,\n  \"email\": \"alice@example.com\"\n}\n```"

	user, err := Parse[TestUser](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if user.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", user.Name)
	}
}

func TestParseUnquotedKeys(t *testing.T) {
	input := `{name: "Bob", age: 28, email: "bob@example.com"}`

	// First fix the JSON
	fixed, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON failed: %v", err)
	}

	user, err := Parse[TestUser](fixed)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if user.Name != "Bob" {
		t.Errorf("Expected name 'Bob', got '%s'", user.Name)
	}
}

func TestParseStringToInt(t *testing.T) {
	input := `{"name": "Charlie", "age": "35", "email": "charlie@example.com"}`

	user, err := Parse[TestUser](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if user.Age != 35 {
		t.Errorf("Expected age 35, got %d", user.Age)
	}
}

func TestParseWithScore(t *testing.T) {
	input := `{"name": "David", "age": "40", "email": "david@example.com"}`

	user, score, err := ParseWithScore[TestUser](input)
	if err != nil {
		t.Fatalf("ParseWithScore failed: %v", err)
	}

	if user.Name != "David" {
		t.Errorf("Expected name 'David', got '%s'", user.Name)
	}

	if score == nil {
		t.Errorf("Expected score, got nil")
	}
}

func TestParseTrailingComma(t *testing.T) {
	input := `{"name": "Eve", "age": 29, "email": "eve@example.com",}`

	// Fix JSON with trailing comma
	fixed, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON failed: %v", err)
	}

	user, err := Parse[TestUser](fixed)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if user.Name != "Eve" {
		t.Errorf("Expected name 'Eve', got '%s'", user.Name)
	}
}

func TestParseBooleanCoercion(t *testing.T) {
	input := `{"title": "Dev", "experience": ["Go", "Rust"], "active": "yes"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !resume.Active {
		t.Errorf("Expected active to be true, got %v", resume.Active)
	}
}

func TestParseArray(t *testing.T) {
	input := `{"title": "Engineer", "experience": ["Go", "Python", "JavaScript"], "active": true}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(resume.Experience) != 3 {
		t.Errorf("Expected 3 experiences, got %d", len(resume.Experience))
	}

	if resume.Experience[0] != "Go" {
		t.Errorf("Expected first experience 'Go', got '%s'", resume.Experience[0])
	}
}

func TestFixJSONMissingQuotes(t *testing.T) {
	input := `{name: John, age: 30}`
	fixed, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON failed: %v", err)
	}

	// Should be valid JSON now
	var m map[string]interface{}
	err = json.Unmarshal([]byte(fixed), &m)
	if err != nil {
		t.Fatalf("Fixed JSON is invalid: %v", err)
	}
}

func TestFixJSONSingleQuotes(t *testing.T) {
	input := `{'name': 'John', 'age': 30}`
	fixed, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON failed: %v", err)
	}

	var m map[string]interface{}
	err = json.Unmarshal([]byte(fixed), &m)
	if err != nil {
		t.Fatalf("Fixed JSON is invalid: %v", err)
	}
}

func TestFuzzyStringMatching(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		distance int
	}{
		{"Cancelled", "Canceled", 1},
		{"canceled", "Canceled", 1},
		{"CANCELED", "Canceled", 6}, // case difference
	}

	for _, tt := range tests {
		dist := stringDistance(tt.input, tt.expected)
		t.Logf("Distance between '%s' and '%s': %d", tt.input, tt.expected, dist)
	}
}

// Test structs with pointer fields
type TestWithPointer struct {
	Name    string  `json:"name"`
	DueDate *string `json:"due_date,omitempty"`
	Count   *int    `json:"count,omitempty"`
}

type TestNestedWithPointer struct {
	Items []TestWithPointer `json:"items"`
}

func TestParsePointerFieldWithValue(t *testing.T) {
	input := `{"name": "Task 1", "due_date": "2025-01-15", "count": 5}`

	result, err := Parse[TestWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "Task 1" {
		t.Errorf("Expected name 'Task 1', got '%s'", result.Name)
	}
	if result.DueDate == nil {
		t.Fatalf("Expected DueDate to be non-nil")
	}
	if *result.DueDate != "2025-01-15" {
		t.Errorf("Expected due_date '2025-01-15', got '%s'", *result.DueDate)
	}
	if result.Count == nil {
		t.Fatalf("Expected Count to be non-nil")
	}
	if *result.Count != 5 {
		t.Errorf("Expected count 5, got %d", *result.Count)
	}
}

func TestParsePointerFieldWithNull(t *testing.T) {
	input := `{"name": "Task 2", "due_date": null, "count": null}`

	result, err := Parse[TestWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "Task 2" {
		t.Errorf("Expected name 'Task 2', got '%s'", result.Name)
	}
	if result.DueDate != nil {
		t.Errorf("Expected DueDate to be nil, got '%v'", result.DueDate)
	}
	if result.Count != nil {
		t.Errorf("Expected Count to be nil, got '%v'", result.Count)
	}
}

func TestParsePointerFieldMissing(t *testing.T) {
	input := `{"name": "Task 3"}`

	result, err := Parse[TestWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "Task 3" {
		t.Errorf("Expected name 'Task 3', got '%s'", result.Name)
	}
	if result.DueDate != nil {
		t.Errorf("Expected DueDate to be nil, got '%v'", result.DueDate)
	}
}

func TestParseNestedStructWithPointer(t *testing.T) {
	input := `{
		"items": [
			{"name": "Item 1", "due_date": "2025-01-10"},
			{"name": "Item 2", "due_date": null},
			{"name": "Item 3"}
		]
	}`

	result, err := Parse[TestNestedWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(result.Items))
	}

	// Item 1 has due_date
	if result.Items[0].DueDate == nil {
		t.Errorf("Expected Item 1 DueDate to be non-nil")
	} else if *result.Items[0].DueDate != "2025-01-10" {
		t.Errorf("Expected Item 1 due_date '2025-01-10', got '%s'", *result.Items[0].DueDate)
	}

	// Item 2 has null due_date
	if result.Items[1].DueDate != nil {
		t.Errorf("Expected Item 2 DueDate to be nil")
	}

	// Item 3 has missing due_date
	if result.Items[2].DueDate != nil {
		t.Errorf("Expected Item 3 DueDate to be nil")
	}
}

func TestParsePointerWithPreambleText(t *testing.T) {
	input := `Based on the analysis, here is the data:

{"name": "Task from LLM", "due_date": "2025-02-01", "count": 10}

Hope this helps!`

	result, err := Parse[TestWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "Task from LLM" {
		t.Errorf("Expected name 'Task from LLM', got '%s'", result.Name)
	}
	if result.DueDate == nil || *result.DueDate != "2025-02-01" {
		t.Errorf("Expected due_date '2025-02-01'")
	}
}

// --- Tests for markdown stripping ---

func TestParseMarkdownBoldNumber(t *testing.T) {
	input := `{"name": "Bold Age", "age": "**35**", "email": "test@test.com"}`

	user, err := Parse[TestUser](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if user.Age != 35 {
		t.Errorf("Expected age 35, got %d", user.Age)
	}
}

func TestParseMarkdownItalicNumber(t *testing.T) {
	input := `{"name": "Italic Age", "age": "_30_", "email": "test@test.com"}`

	user, err := Parse[TestUser](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if user.Age != 30 {
		t.Errorf("Expected age 30, got %d", user.Age)
	}
}

// --- Tests for numbers with units ---

type TestMeasurement struct {
	Label    string  `json:"label"`
	Value    int     `json:"value"`
	Precise  float64 `json:"precise"`
}

func TestParseNumberWithTrailingUnit(t *testing.T) {
	input := `{"label": "age", "value": "30 years", "precise": 0}`

	result, err := Parse[TestMeasurement](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result.Value != 30 {
		t.Errorf("Expected value 30, got %d", result.Value)
	}
}

func TestParseNumberWithKSuffix(t *testing.T) {
	input := `{"label": "salary", "value": "200K", "precise": 0}`

	result, err := Parse[TestMeasurement](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result.Value != 200000 {
		t.Errorf("Expected value 200000, got %d", result.Value)
	}
}

func TestParseNumberWithMSuffix(t *testing.T) {
	input := `{"label": "revenue", "precise": "2.5M", "value": 0}`

	result, err := Parse[TestMeasurement](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result.Precise != 2500000 {
		t.Errorf("Expected precise 2500000, got %f", result.Precise)
	}
}

func TestParseNumberWithCurrencyAndUnit(t *testing.T) {
	input := `{"label": "price", "value": "$200K", "precise": 0}`

	result, err := Parse[TestMeasurement](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result.Value != 200000 {
		t.Errorf("Expected value 200000, got %d", result.Value)
	}
}

func TestParseNumberWithGBUnit(t *testing.T) {
	input := `{"label": "storage", "value": "4 GB", "precise": 0}`

	result, err := Parse[TestMeasurement](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result.Value != 4 {
		t.Errorf("Expected value 4, got %d", result.Value)
	}
}

// --- Tests for null string variants ---

func TestParseNullStringNA(t *testing.T) {
	input := `{"name": "Null Test", "due_date": "N/A", "count": null}`

	result, err := Parse[TestWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result.DueDate != nil {
		t.Errorf("Expected DueDate to be nil for 'N/A', got '%v'", *result.DueDate)
	}
}

func TestParseNullStringNone(t *testing.T) {
	input := `{"name": "None Test", "due_date": "none"}`

	result, err := Parse[TestWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result.DueDate != nil {
		t.Errorf("Expected DueDate to be nil for 'none', got '%v'", *result.DueDate)
	}
}

func TestParseNullStringUnknown(t *testing.T) {
	input := `{"name": "Unknown Test", "due_date": "unknown"}`

	result, err := Parse[TestWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result.DueDate != nil {
		t.Errorf("Expected DueDate to be nil for 'unknown', got '%v'", *result.DueDate)
	}
}

// --- Tests for comma-separated string to slice ---

func TestParseCommaSeparatedStringToSlice(t *testing.T) {
	input := `{"title": "Dev", "experience": "python, go, rust", "active": true}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(resume.Experience) != 3 {
		t.Fatalf("Expected 3 experiences, got %d", len(resume.Experience))
	}
	if resume.Experience[0] != "python" {
		t.Errorf("Expected first experience 'python', got '%s'", resume.Experience[0])
	}
	if resume.Experience[1] != "go" {
		t.Errorf("Expected second experience 'go', got '%s'", resume.Experience[1])
	}
	if resume.Experience[2] != "rust" {
		t.Errorf("Expected third experience 'rust', got '%s'", resume.Experience[2])
	}
}

// --- Tests for embedded struct support ---

type TestAddress struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

type TestPerson struct {
	TestAddress
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestParseEmbeddedStruct(t *testing.T) {
	input := `{"name": "Alice", "age": 30, "city": "NYC", "country": "US"}`

	person, err := Parse[TestPerson](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if person.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", person.Name)
	}
	if person.Age != 30 {
		t.Errorf("Expected age 30, got %d", person.Age)
	}
	if person.City != "NYC" {
		t.Errorf("Expected city 'NYC', got '%s'", person.City)
	}
	if person.Country != "US" {
		t.Errorf("Expected country 'US', got '%s'", person.Country)
	}
}

type TestMeta struct {
	CreatedBy string `json:"created_by"`
}

type TestDocument struct {
	TestMeta
	Title   string `json:"title"`
	Content string `json:"content"`
}

func TestParseEmbeddedStructNested(t *testing.T) {
	input := `{"title": "Doc", "content": "Hello", "created_by": "admin"}`

	doc, err := Parse[TestDocument](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if doc.Title != "Doc" {
		t.Errorf("Expected title 'Doc', got '%s'", doc.Title)
	}
	if doc.CreatedBy != "admin" {
		t.Errorf("Expected created_by 'admin', got '%s'", doc.CreatedBy)
	}
}

// --- Tests for extended boolean variants ---

func TestParseBooleanY(t *testing.T) {
	input := `{"title": "Dev", "experience": ["Go"], "active": "y"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !resume.Active {
		t.Errorf("Expected active to be true for 'y'")
	}
}

func TestParseBooleanN(t *testing.T) {
	input := `{"title": "Dev", "experience": ["Go"], "active": "n"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if resume.Active {
		t.Errorf("Expected active to be false for 'n'")
	}
}

func TestParseBooleanEnabled(t *testing.T) {
	input := `{"title": "Dev", "experience": ["Go"], "active": "enabled"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !resume.Active {
		t.Errorf("Expected active to be true for 'enabled'")
	}
}

func TestParseBooleanDisabled(t *testing.T) {
	input := `{"title": "Dev", "experience": ["Go"], "active": "disabled"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if resume.Active {
		t.Errorf("Expected active to be false for 'disabled'")
	}
}

func TestParseBooleanActive(t *testing.T) {
	input := `{"title": "Dev", "experience": ["Go"], "active": "active"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !resume.Active {
		t.Errorf("Expected active to be true for 'active'")
	}
}

func TestParseBooleanInactive(t *testing.T) {
	input := `{"title": "Dev", "experience": ["Go"], "active": "inactive"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if resume.Active {
		t.Errorf("Expected active to be false for 'inactive'")
	}
}

// --- Tests for ParsePartial ---

func TestParsePartialValidJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  TestUser
	}{
		{
			name:  "complete object",
			input: `{"name": "Alice", "age": 30, "email": "alice@test.com"}`,
			want:  TestUser{Name: "Alice", Age: 30, Email: "alice@test.com"},
		},
		{
			name:  "with type coercion",
			input: `{"name": "Bob", "age": "25", "email": "bob@test.com"}`,
			want:  TestUser{Name: "Bob", Age: 25, Email: "bob@test.com"},
		},
		{
			name:  "embedded in markdown",
			input: "Here is the data:\n```json\n{\"name\": \"Eve\", \"age\": 28, \"email\": \"eve@test.com\"}\n```",
			want:  TestUser{Name: "Eve", Age: 28, Email: "eve@test.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, state, err := ParsePartial[TestUser](tt.input)
			if err != nil {
				t.Fatalf("ParsePartial failed: %v", err)
			}
			if state != Complete {
				t.Errorf("Expected CompletionState Complete, got %v", state)
			}
			if result.Name != tt.want.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tt.want.Name)
			}
			if result.Age != tt.want.Age {
				t.Errorf("Age: got %d, want %d", result.Age, tt.want.Age)
			}
			if result.Email != tt.want.Email {
				t.Errorf("Email: got %q, want %q", result.Email, tt.want.Email)
			}
		})
	}
}

func TestParsePartialInvalidJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "no JSON at all",
			input: "this is just plain text with no JSON",
		},
		{
			name:  "empty string",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, state, err := ParsePartial[TestUser](tt.input)
			if err == nil {
				t.Fatalf("Expected error for invalid input %q, got result: %+v", tt.input, result)
			}
			// On error, ParsePartial returns Complete as the state
			if state != Complete {
				t.Errorf("Expected CompletionState Complete on error, got %v", state)
			}
			// Result should be zero value
			if result != (TestUser{}) {
				t.Errorf("Expected zero TestUser on error, got %+v", result)
			}
		})
	}
}

// --- Tests for InstructorParser configuration ---

func TestInstructorParserWithStrict(t *testing.T) {
	t.Run("strict rejects malformed JSON", func(t *testing.T) {
		parser := NewInstructorParser().WithStrict(true)

		var user TestUser
		err := parser.Unmarshal([]byte(`{name: "Alice", age: 30}`), &user)
		if err == nil {
			t.Fatal("Expected strict mode to reject unquoted keys, got nil error")
		}
	})

	t.Run("strict accepts valid JSON", func(t *testing.T) {
		parser := NewInstructorParser().WithStrict(true)

		var user TestUser
		err := parser.Unmarshal(
			[]byte(`{"name": "Alice", "age": 30, "email": "alice@test.com"}`),
			&user,
		)
		if err != nil {
			t.Fatalf("Expected valid JSON to succeed in strict mode: %v", err)
		}
		if user.Name != "Alice" {
			t.Errorf("Name: got %q, want %q", user.Name, "Alice")
		}
		if user.Age != 30 {
			t.Errorf("Age: got %d, want %d", user.Age, 30)
		}
	})

	t.Run("chaining returns same parser", func(t *testing.T) {
		parser := NewInstructorParser()
		returned := parser.WithStrict(true)
		if returned != parser {
			t.Error("WithStrict should return the same InstructorParser for chaining")
		}
	})
}

func TestInstructorParserWithIncompleteJSON(t *testing.T) {
	t.Run("chaining returns same parser", func(t *testing.T) {
		parser := NewInstructorParser()
		returned := parser.WithIncompleteJSON(true)
		if returned != parser {
			t.Error("WithIncompleteJSON should return the same InstructorParser for chaining")
		}
	})

	t.Run("complete JSON still parses after enabling", func(t *testing.T) {
		parser := NewInstructorParser().WithIncompleteJSON(true)

		var user TestUser
		err := parser.Unmarshal(
			[]byte(`{"name": "Bob", "age": 40, "email": "bob@test.com"}`),
			&user,
		)
		if err != nil {
			t.Fatalf("Expected complete JSON to parse with WithIncompleteJSON enabled: %v", err)
		}
		if user.Name != "Bob" {
			t.Errorf("Name: got %q, want %q", user.Name, "Bob")
		}
	})

	t.Run("chaining both options", func(t *testing.T) {
		parser := NewInstructorParser().WithStrict(false).WithIncompleteJSON(true)

		var resume TestResume
		err := parser.Unmarshal(
			[]byte(`{"title": "Engineer", "experience": ["Go"], "active": true}`),
			&resume,
		)
		if err != nil {
			t.Fatalf("Chained configuration should parse valid JSON: %v", err)
		}
		if resume.Title != "Engineer" {
			t.Errorf("Title: got %q, want %q", resume.Title, "Engineer")
		}
	})
}

func TestInstructorParserUnmarshalNil(t *testing.T) {
	parser := NewInstructorParser()
	err := parser.Unmarshal([]byte(`{"name": "test"}`), nil)
	if err != nil {
		t.Errorf("Unmarshal with nil target should return nil error, got: %v", err)
	}
}

// --- Integration edge cases ---

// TestUintFields is a struct for testing unsigned integer coercion.
type TestUintFields struct {
	Count   uint   `json:"count"`
	Flags   uint32 `json:"flags"`
	BigID   uint64 `json:"big_id"`
}

func TestParseUintField(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TestUintFields
	}{
		{
			name:  "direct numeric values",
			input: `{"count": 42, "flags": 255, "big_id": 9999999}`,
			want:  TestUintFields{Count: 42, Flags: 255, BigID: 9999999},
		},
		{
			name:  "string to uint coercion",
			input: `{"count": "10", "flags": "128", "big_id": "5000"}`,
			want:  TestUintFields{Count: 10, Flags: 128, BigID: 5000},
		},
		{
			name:  "zero values",
			input: `{"count": 0, "flags": 0, "big_id": 0}`,
			want:  TestUintFields{Count: 0, Flags: 0, BigID: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse[TestUintFields](tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if result.Count != tt.want.Count {
				t.Errorf("Count: got %d, want %d", result.Count, tt.want.Count)
			}
			if result.Flags != tt.want.Flags {
				t.Errorf("Flags: got %d, want %d", result.Flags, tt.want.Flags)
			}
			if result.BigID != tt.want.BigID {
				t.Errorf("BigID: got %d, want %d", result.BigID, tt.want.BigID)
			}
		})
	}
}

// TestArrayFields is a struct for testing fixed-size array coercion.
type TestArrayFields struct {
	Tags   [3]string `json:"tags"`
	Scores [2]int    `json:"scores"`
}

func TestParseArrayField(t *testing.T) {
	tests := []struct {
		name  string
		input string
		tags  [3]string
		scores [2]int
	}{
		{
			name:   "exact length match",
			input:  `{"tags": ["go", "rust", "python"], "scores": [95, 87]}`,
			tags:   [3]string{"go", "rust", "python"},
			scores: [2]int{95, 87},
		},
		{
			name:   "fewer elements than array size",
			input:  `{"tags": ["go"], "scores": [100]}`,
			tags:   [3]string{"go", "", ""},
			scores: [2]int{100, 0},
		},
		{
			name:   "more elements than array size truncates",
			input:  `{"tags": ["a", "b", "c", "d", "e"], "scores": [1, 2, 3]}`,
			tags:   [3]string{"a", "b", "c"},
			scores: [2]int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse[TestArrayFields](tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if result.Tags != tt.tags {
				t.Errorf("Tags: got %v, want %v", result.Tags, tt.tags)
			}
			if result.Scores != tt.scores {
				t.Errorf("Scores: got %v, want %v", result.Scores, tt.scores)
			}
		})
	}
}

// TestStringFromNumber is a struct for testing number-to-string coercion.
type TestStringFromNumber struct {
	Count string `json:"count"`
	Label string `json:"label"`
}

func TestParseStringFromNumber(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount string
		wantLabel string
	}{
		{
			name:      "integer to string",
			input:     `{"count": 42, "label": "test"}`,
			wantCount: "42",
			wantLabel: "test",
		},
		{
			name:      "float to string",
			input:     `{"count": 3.14, "label": "pi"}`,
			wantCount: "3.14",
			wantLabel: "pi",
		},
		{
			name:      "boolean to string",
			input:     `{"count": true, "label": "flag"}`,
			wantCount: "true",
			wantLabel: "flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse[TestStringFromNumber](tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if result.Count != tt.wantCount {
				t.Errorf("Count: got %q, want %q", result.Count, tt.wantCount)
			}
			if result.Label != tt.wantLabel {
				t.Errorf("Label: got %q, want %q", result.Label, tt.wantLabel)
			}
		})
	}
}

func TestParseMapTarget(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
	}{
		{
			name:     "simple flat object",
			input:    `{"name": "Alice", "role": "engineer"}`,
			wantKeys: []string{"name", "role"},
		},
		{
			name:     "mixed value types",
			input:    `{"count": 42, "active": true, "label": "test"}`,
			wantKeys: []string{"count", "active", "label"},
		},
	}

	mapType := reflect.TypeOf(map[string]interface{}{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			result, err := parser.Parse(tt.input, mapType)
			if err != nil {
				t.Fatalf("Parse to map failed: %v", err)
			}
			m, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected map[string]interface{}, got %T", result)
			}
			for _, key := range tt.wantKeys {
				if _, exists := m[key]; !exists {
					t.Errorf("Expected key %q in map, keys present: %v", key, mapKeys(m))
				}
			}
		})
	}
}

// mapKeys returns the keys of a map for diagnostic output.
func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestNestedPointerStruct is a struct for testing pointer-to-struct fields.
type TestNestedPointerStruct struct {
	Owner   *TestUser `json:"owner"`
	Project string    `json:"project"`
}

func TestParseNestedPointerToStruct(t *testing.T) {
	t.Run("populated pointer", func(t *testing.T) {
		input := `{"project": "gsap", "owner": {"name": "Alice", "age": 30, "email": "alice@test.com"}}`
		result, err := Parse[TestNestedPointerStruct](input)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if result.Project != "gsap" {
			t.Errorf("Project: got %q, want %q", result.Project, "gsap")
		}
		if result.Owner == nil {
			t.Fatal("Expected Owner to be non-nil")
		}
		if result.Owner.Name != "Alice" {
			t.Errorf("Owner.Name: got %q, want %q", result.Owner.Name, "Alice")
		}
		if result.Owner.Age != 30 {
			t.Errorf("Owner.Age: got %d, want %d", result.Owner.Age, 30)
		}
	})

	t.Run("null pointer", func(t *testing.T) {
		input := `{"project": "gsap", "owner": null}`
		result, err := Parse[TestNestedPointerStruct](input)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if result.Owner != nil {
			t.Errorf("Expected Owner to be nil for null JSON value, got %+v", result.Owner)
		}
	})

	t.Run("missing pointer field", func(t *testing.T) {
		input := `{"project": "gsap"}`
		result, err := Parse[TestNestedPointerStruct](input)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if result.Owner != nil {
			t.Errorf("Expected Owner to be nil for missing field, got %+v", result.Owner)
		}
	})
}

func TestParseScoreFlags(t *testing.T) {
	t.Run("string-to-int coercion sets flag", func(t *testing.T) {
		input := `{"name": "Test", "age": "25", "email": "t@t.com"}`
		_, score, err := ParseWithScore[TestUser](input)
		if err != nil {
			t.Fatalf("ParseWithScore failed: %v", err)
		}
		flags := score.Flags()
		if _, ok := flags[FlagStringToInt]; !ok {
			t.Errorf("Expected %s flag to be set, flags: %v", FlagStringToInt, flags)
		}
	})

	t.Run("string-to-bool coercion sets flag", func(t *testing.T) {
		input := `{"title": "Dev", "experience": ["Go"], "active": "yes"}`
		_, score, err := ParseWithScore[TestResume](input)
		if err != nil {
			t.Fatalf("ParseWithScore failed: %v", err)
		}
		flags := score.Flags()
		if _, ok := flags[FlagStringToBool]; !ok {
			t.Errorf("Expected %s flag to be set, flags: %v", FlagStringToBool, flags)
		}
	})

	t.Run("clean parse has zero score", func(t *testing.T) {
		input := `{"name": "Clean", "age": 30, "email": "c@c.com"}`
		_, score, err := ParseWithScore[TestUser](input)
		if err != nil {
			t.Fatalf("ParseWithScore failed: %v", err)
		}
		if score.Total() != 0 {
			t.Errorf("Expected zero score for clean JSON, got %d (flags: %v)", score.Total(), score.Flags())
		}
	})

	t.Run("embedded struct sets flag", func(t *testing.T) {
		input := `{"name": "Alice", "age": 30, "city": "NYC", "country": "US"}`
		_, score, err := ParseWithScore[TestPerson](input)
		if err != nil {
			t.Fatalf("ParseWithScore failed: %v", err)
		}
		flags := score.Flags()
		if _, ok := flags[FlagEmbeddedStruct]; !ok {
			t.Errorf("Expected %s flag to be set, flags: %v", FlagEmbeddedStruct, flags)
		}
	})

	t.Run("comma-split sets flag", func(t *testing.T) {
		input := `{"title": "Dev", "experience": "go, rust", "active": true}`
		_, score, err := ParseWithScore[TestResume](input)
		if err != nil {
			t.Fatalf("ParseWithScore failed: %v", err)
		}
		flags := score.Flags()
		if _, ok := flags[FlagCommaSplitToSlice]; !ok {
			t.Errorf("Expected %s flag to be set, flags: %v", FlagCommaSplitToSlice, flags)
		}
	})

	t.Run("multiple coercions accumulate score", func(t *testing.T) {
		input := `{"name": "Multi", "age": "30", "email": "m@m.com"}`
		_, score, err := ParseWithScore[TestUser](input)
		if err != nil {
			t.Fatalf("ParseWithScore failed: %v", err)
		}
		if score.Total() <= 0 {
			t.Errorf("Expected positive score for coerced JSON, got %d", score.Total())
		}
	})
}
