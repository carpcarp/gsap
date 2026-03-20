package sap

import (
	"encoding/json"
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
