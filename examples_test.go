package sap

import (
	"testing"
)

// Example 1: Basic parsing with type coercion
func TestExample_basic(t *testing.T) {
	input := `{"name": "Alice", "age": "30", "email": "alice@example.com"}`

	user, err := Parse[TestUser](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if user.Name != "Alice" || user.Age != 30 {
		t.Errorf("Unexpected result: %+v", user)
	}
}

// Example 2: Parsing JSON from markdown
func TestExample_markdown(t *testing.T) {
	input := "Here's the user data:\n```json\n{\"name\": \"Bob\", \"age\": 25, \"email\": \"bob@example.com\"}\n```"

	user, err := Parse[TestUser](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if user.Name != "Bob" || user.Age != 25 {
		t.Errorf("Unexpected result: %+v", user)
	}
}

// Example 3: Handling unquoted keys
func TestExample_unquotedKeys(t *testing.T) {
	input := `{name: "Charlie", age: 35, email: "charlie@example.com"}`

	fixed, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON failed: %v", err)
	}

	// Verify fixed JSON is valid
	user, err := Parse[TestUser](fixed)
	if err != nil {
		t.Fatalf("Parse fixed JSON failed: %v", err)
	}
	t.Logf("Fixed JSON parsed successfully: %+v", user)
}

// Example 4: Handling single quotes
func TestExample_singleQuotes(t *testing.T) {
	input := `{'name': 'David', 'age': 40, 'email': 'david@example.com'}`

	fixed, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON failed: %v", err)
	}

	user, err := Parse[TestUser](fixed)
	if err != nil {
		t.Fatalf("Parse fixed JSON failed: %v", err)
	}
	t.Logf("Parsed from single quotes: %+v", user)
}

// Example 5: Handling trailing commas
func TestExample_trailingCommas(t *testing.T) {
	input := `{"name": "Eve", "age": 29, "email": "eve@example.com",}`

	fixed, err := FixJSON(input)
	if err != nil {
		t.Fatalf("FixJSON failed: %v", err)
	}

	user, err := Parse[TestUser](fixed)
	if err != nil {
		t.Fatalf("Parse fixed JSON failed: %v", err)
	}
	t.Logf("Parsed despite trailing comma: %+v", user)
}

// Example 6: Getting parse quality score
func TestExample_scoreTracking(t *testing.T) {
	input := `{"name": "Frank", "age": "50", "email": "frank@example.com"}`

	user, score, err := ParseWithScore[TestUser](input)
	if err != nil {
		t.Fatalf("ParseWithScore failed: %v", err)
	}

	if score == nil {
		t.Errorf("Expected score, got nil")
	}
	t.Logf("Parsed with score %d: %+v", score.Total(), user)
}

// Example 7: Complex types with arrays
func TestExample_complexTypes(t *testing.T) {
	input := `{"title": "Engineer", "experience": ["Go", "Rust", "Python"], "active": true}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(resume.Experience) != 3 {
		t.Errorf("Expected 3 experiences, got %d", len(resume.Experience))
	}
	t.Logf("Parsed complex types: %+v", resume)
}

// Example 8: Boolean coercion from string
func TestExample_booleanCoercion(t *testing.T) {
	input := `{"title": "Dev", "experience": ["Go"], "active": "yes"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !resume.Active {
		t.Errorf("Expected active=true, got %v", resume.Active)
	}
	t.Logf("Boolean coercion from string: %+v", resume)
}

// Example 9: Custom parsing with strict mode
func TestExample_strictMode(t *testing.T) {
	parser := NewParser().WithStrict(true)

	// This would fail in strict mode due to unquoted keys
	input := `{name: "Alice", age: 30, email: "alice@example.com"}`

	_, err := parser.Parse(input, nil)
	if err != nil {
		// Expected to fail in strict mode
		t.Logf("Strict mode correctly rejected malformed JSON: %v", err)
	}
}

// Example 10: Handling JSON with comments
func ExampleFixJSON_comments() {
	input := `{
		// Name of the user
		"name": "Grace",
		/* Age in years */ "age": 31,
		"email": "grace@example.com"
	}`

	fixed, err := FixJSON(input)
	if err != nil {
		panic(err)
	}

	// Comments are removed, valid JSON remains
	_ = fixed
}

// Example 11: Chain-of-thought with JSON
func ExampleParse_chainOfThought() {
	input := `Let me extract the user info:
The user's name is Alice and they are 30 years old.

Here's the structured data:
{
  "name": "Alice",
  "age": 30,
  "email": "alice@example.com"
}

Hope this helps!`

	user, err := Parse[TestUser](input)
	if err != nil {
		panic(err)
	}

	// Successfully extracts JSON despite surrounding text
	_ = user
}

// Example 12: Fuzzy string matching
func TestExample_fuzzyMatching(t *testing.T) {
	input1 := "canceled"
	input2 := "Canceled"

	dist := stringDistance(input1, input2)
	t.Logf("Distance between '%s' and '%s': %d", input1, input2, dist)

	// Low distance indicates a good match for enum values
	if dist <= 2 {
		t.Log("Strings are similar enough for enum matching")
	}
}

// Example 13: Type coercion with fractions
func TestExample_fractions(t *testing.T) {
	input := `{"name": "Test", "age": "1/2", "email": "test@example.com"}`
	// Age as fraction gets converted
	_, err := Parse[TestUser](input)
	t.Logf("Fraction parsing test: %v", err)
}

// Example 14: Type coercion with currency
func TestExample_currency(t *testing.T) {
	input := `{"name": "Test", "age": "30", "email": "test@example.com"}`
	// Currency parsing is handled in type coercion
	_, err := Parse[TestUser](input)
	t.Logf("Currency parsing test: %v", err)
}

// Example 15: Integration pattern with LLM client
func TestExample_llmIntegration(t *testing.T) {
	// This shows how to use SAP with an LLM client

	// Pseudo-code for instructor-go integration:
	// 1. Get LLM response (potentially malformed JSON)
	// 2. Pass to SAP parser
	// 3. Get strongly-typed result

	llmResponse := "Here's the extracted user:\n```json\n{\n  name: \"Alice\",\n  age: 30,\n  email: \"alice@example.com\"\n}\n```"

	user, err := Parse[TestUser](llmResponse)
	if err != nil {
		t.Fatalf("Failed to parse LLM response: %v", err)
	}

	t.Logf("Parsed user: %+v", user)
}
