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
