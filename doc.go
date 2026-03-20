// Package sap provides a Schema-Aligned Parser for extracting typed data
// from messy LLM-generated JSON.
//
// LLMs frequently produce JSON that is technically invalid: wrapped in
// markdown code blocks, decorated with explanatory text, or containing
// syntax errors like unquoted keys, single quotes, trailing commas, and
// inline comments. Package sap handles all of this automatically. It
// extracts JSON candidates from free-form text, fixes common formatting
// problems, coerces mismatched types to fit your target struct, and picks
// the best parse result via a scoring system.
//
// # Quick Start
//
// The simplest way to use sap is the generic Parse function:
//
//	type User struct {
//	    Name  string `json:"name"`
//	    Age   int    `json:"age"`
//	    Email string `json:"email"`
//	}
//
//	user, err := sap.Parse[User](llmResponse)
//
// If you need to inspect how much coercion was required, use
// ParseWithScore. Lower scores indicate cleaner parses:
//
//	user, score, err := sap.ParseWithScore[User](llmResponse)
//	fmt.Printf("parse quality score: %d\n", score.Total())
//
// To fix malformed JSON without parsing it into a struct, use FixJSON
// directly:
//
//	fixed, err := sap.FixJSON(`{name: 'Alice', age: 30,}`)
//	// fixed is now valid JSON: {"name": "Alice", "age": 30}
//
// # JSON Extraction
//
// When the input is not already valid JSON, the extractor runs through
// several strategies in order:
//
//  1. Try parsing the trimmed input as standard JSON.
//  2. Look for JSON inside markdown code blocks (```json ... ```).
//  3. Scan the text for balanced { } and [ ] blocks.
//  4. Attempt to fix any candidate that fails standard parsing by
//     quoting unquoted keys, converting single quotes and backticks to
//     double quotes, removing trailing commas, and stripping comments.
//
// The best candidate is selected by parsing each one against your target
// type and comparing scores.
//
// # Type Coercion
//
// The type coercer transforms JSON values to fit your Go struct fields.
// Supported coercions include:
//
//   - String to int/float: "42" becomes 42, "$1,234.56" becomes 1234.56
//   - String to bool: "true", "yes", "1", "on" become true;
//     "false", "no", "0", "off" become false
//   - Fractions: "1/5" becomes 0.2
//   - Currency: "$1,000" becomes 1000
//   - Number to bool: 0 is false, nonzero is true
//   - Bool to int: true is 1, false is 0
//   - Case-insensitive struct field matching
//   - Fuzzy enum matching with Unicode normalization
//
// Each coercion adds a penalty to the parse score so you can distinguish
// a clean parse from one that required significant transformation.
//
// # Streaming
//
// For streaming LLM responses, ParsePartial accepts incomplete JSON and
// reports a CompletionState (Complete, Incomplete, or Pending):
//
//	user, state, err := sap.ParsePartial[User](partialResponse)
//
// # instructor-go Integration
//
// To use sap as the parser for instructor-go, create an InstructorParser:
//
//	parser := sap.NewInstructorParser()
//	client := instructor.FromOpenAI(openaiClient,
//	    instructor.WithParser(parser),
//	)
//
// The InstructorParser implements the instructor-go Parser interface,
// replacing its default JSON unmarshaling with the full sap extraction
// and coercion pipeline.
package sap
