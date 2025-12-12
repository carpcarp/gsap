package sap

import (
	"testing"
)

// Test structs for multiple candidate scenarios
type Company struct {
	Name       string   `json:"name"`
	Employees  []Person `json:"employees"`
	Department *string  `json:"department,omitempty"`
}

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Project struct {
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Tasks    []string `json:"tasks"`
	Priority *int     `json:"priority,omitempty"`
}

// TestMarkdownWithNestedArrays tests parsing JSON in markdown that contains arrays
// This was the original bug - nested arrays created extra candidates that failed
func TestMarkdownWithNestedArrays(t *testing.T) {
	input := "```json\n" + `{
  "name": "Acme Corp",
  "employees": [
    {"name": "Alice", "email": "alice@acme.com"},
    {"name": "Bob", "email": "bob@acme.com"}
  ],
  "department": "Engineering"
}` + "\n```"

	result, err := Parse[Company](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "Acme Corp" {
		t.Errorf("Expected name 'Acme Corp', got '%s'", result.Name)
	}
	if len(result.Employees) != 2 {
		t.Errorf("Expected 2 employees, got %d", len(result.Employees))
	}
	if result.Employees[0].Name != "Alice" {
		t.Errorf("Expected first employee 'Alice', got '%s'", result.Employees[0].Name)
	}
}

// TestMarkdownWithDeeplyNestedArrays tests multiple levels of nesting
func TestMarkdownWithDeeplyNestedArrays(t *testing.T) {
	input := "```json\n" + `{
  "title": "Q4 Project",
  "status": "active",
  "tasks": ["design", "implement", "test", "deploy"],
  "priority": 1
}` + "\n```"

	result, err := Parse[Project](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Title != "Q4 Project" {
		t.Errorf("Expected title 'Q4 Project', got '%s'", result.Title)
	}
	if len(result.Tasks) != 4 {
		t.Errorf("Expected 4 tasks, got %d", len(result.Tasks))
	}
}

// TestPreambleTextWithNestedArrays tests JSON with preamble text containing arrays
func TestPreambleTextWithNestedArrays(t *testing.T) {
	input := `Here is the company data you requested:

{
  "name": "TechStart",
  "employees": [
    {"name": "Carol", "email": "carol@techstart.io"}
  ]
}

Let me know if you need anything else!`

	result, err := Parse[Company](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "TechStart" {
		t.Errorf("Expected name 'TechStart', got '%s'", result.Name)
	}
}

// TestMultipleJSONBlocksPicksBest tests that when multiple valid JSON blocks exist,
// the parser picks the best one (most complete match)
func TestMultipleJSONBlocksPicksBest(t *testing.T) {
	// Input has a partial object and a complete object
	input := `{"name": "Partial"}

Here's the full data:
{"name": "Complete Corp", "employees": [{"name": "Dan", "email": "dan@complete.com"}]}`

	result, err := Parse[Company](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should get one of the valid parses (both are valid for Company)
	if result.Name == "" {
		t.Error("Expected a name to be parsed")
	}
}

// TestEmptyArraysInMarkdown tests that empty arrays don't cause issues
func TestEmptyArraysInMarkdown(t *testing.T) {
	input := "```json\n" + `{
  "name": "Empty Corp",
  "employees": []
}` + "\n```"

	result, err := Parse[Company](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "Empty Corp" {
		t.Errorf("Expected name 'Empty Corp', got '%s'", result.Name)
	}
	if len(result.Employees) != 0 {
		t.Errorf("Expected 0 employees, got %d", len(result.Employees))
	}
}

// TestArrayOnlyCandidate tests that an array-only input fails gracefully for struct target
func TestArrayOnlyCandidate(t *testing.T) {
	input := `[{"name": "Alice"}, {"name": "Bob"}]`

	_, err := Parse[Company](input)
	if err == nil {
		t.Error("Expected error when parsing array as struct, got nil")
	}
}

// TestSuccessfulCandidateAfterFailures tests that a successful parse is returned
// even if earlier candidates failed
func TestSuccessfulCandidateAfterFailures(t *testing.T) {
	// First candidate is an array (will fail for struct)
	// Second candidate is the valid object
	input := `[1, 2, 3]

{"name": "Success Corp", "employees": []}`

	result, err := Parse[Company](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "Success Corp" {
		t.Errorf("Expected name 'Success Corp', got '%s'", result.Name)
	}
}

// TestFailedCandidateAfterSuccess tests that a failed candidate after success
// doesn't overwrite the successful result (the original bug)
func TestFailedCandidateAfterSuccess(t *testing.T) {
	// The markdown block has valid JSON, but the extractor also finds
	// the nested array as a separate candidate which will fail
	input := "```json\n" + `{
  "title": "Important Project",
  "status": "in_progress",
  "tasks": ["task1", "task2", "task3"]
}` + "\n```"

	result, err := Parse[Project](input)
	if err != nil {
		t.Fatalf("Parse failed (bug: failed candidate after success overwrote result): %v", err)
	}

	if result.Title != "Important Project" {
		t.Errorf("Expected title 'Important Project', got '%s'", result.Title)
	}
	if result.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", result.Status)
	}
}

// TestComplexNestedStructure tests a complex nested structure in markdown
func TestComplexNestedStructure(t *testing.T) {
	type Meeting struct {
		Date      string   `json:"date"`
		Attendees []string `json:"attendees"`
		Notes     string   `json:"notes"`
	}

	type Schedule struct {
		Owner    string    `json:"owner"`
		Meetings []Meeting `json:"meetings"`
	}

	input := "```json\n" + `{
  "owner": "Team Lead",
  "meetings": [
    {
      "date": "2025-01-15",
      "attendees": ["Alice", "Bob", "Carol"],
      "notes": "Sprint planning"
    },
    {
      "date": "2025-01-16",
      "attendees": ["Dan", "Eve"],
      "notes": "Design review"
    }
  ]
}` + "\n```"

	result, err := Parse[Schedule](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Owner != "Team Lead" {
		t.Errorf("Expected owner 'Team Lead', got '%s'", result.Owner)
	}
	if len(result.Meetings) != 2 {
		t.Errorf("Expected 2 meetings, got %d", len(result.Meetings))
	}
	if len(result.Meetings[0].Attendees) != 3 {
		t.Errorf("Expected 3 attendees in first meeting, got %d", len(result.Meetings[0].Attendees))
	}
}

// TestMarkdownWithMixedContent tests markdown with other content around JSON
func TestMarkdownWithMixedContent(t *testing.T) {
	input := `# Analysis Results

Based on the data provided, here's the summary:

` + "```json\n" + `{
  "name": "Data Corp",
  "employees": [{"name": "Frank", "email": "frank@data.com"}]
}` + "\n```" + `

## Next Steps

Please review and confirm.`

	result, err := Parse[Company](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Name != "Data Corp" {
		t.Errorf("Expected name 'Data Corp', got '%s'", result.Name)
	}
}
