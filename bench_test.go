package sap

import (
	"testing"
)

func BenchmarkParseValidJSON(b *testing.B) {
	input := `{"name": "Alice Johnson", "age": 30, "email": "alice.johnson@example.com"}`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse[TestUser](input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseMarkdown(b *testing.B) {
	input := `Based on my analysis of the user's profile, here is the structured data:

` + "```json" + `
{
  "name": "Alice Johnson",
  "age": 30,
  "email": "alice.johnson@example.com"
}
` + "```" + `

Let me know if you need anything else!`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse[TestUser](input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseMalformed(b *testing.B) {
	// Unquoted keys, single quotes, trailing comma — typical LLM output
	input := `{name: 'Alice Johnson', age: 30, email: 'alice@example.com',}`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixed, err := FixJSON(input)
		if err != nil {
			b.Fatal(err)
		}
		_, err = Parse[TestUser](fixed)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseCoercion(b *testing.B) {
	// String-to-int and string-to-bool coercion
	input := `{"name": "Alice Johnson", "age": "30", "email": "alice@example.com"}`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user, err := Parse[TestUser](input)
		if err != nil {
			b.Fatal(err)
		}
		if user.Age != 30 {
			b.Fatalf("expected age 30, got %d", user.Age)
		}
	}
}

type BenchNestedProject struct {
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Active       bool               `json:"active"`
	Contributors []BenchContributor `json:"contributors"`
	Tags         []string           `json:"tags"`
}

type BenchContributor struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Commits int   `json:"commits"`
}

func BenchmarkParseNestedStruct(b *testing.B) {
	input := `{
		"name": "gsap",
		"version": "1.0.0",
		"active": true,
		"contributors": [
			{"name": "Alice", "email": "alice@example.com", "commits": 142},
			{"name": "Bob", "email": "bob@example.com", "commits": 87},
			{"name": "Charlie", "email": "charlie@example.com", "commits": 53}
		],
		"tags": ["go", "json", "llm", "parser", "ai"]
	}`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse[BenchNestedProject](input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFixJSON(b *testing.B) {
	// A mix of common LLM JSON issues: unquoted keys, single quotes,
	// trailing commas, and comments
	input := `{
		// User profile data
		name: 'Alice Johnson',
		age: 30,
		email: 'alice@example.com',
		'active': true,
	}`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FixJSON(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExtractJSON(b *testing.B) {
	// Realistic LLM response with chain-of-thought reasoning surrounding JSON
	input := `I've analyzed the user's information carefully. Based on the context provided,
the user appears to be a software engineer with significant experience.

Here is the extracted structured data:

{
  "name": "Alice Johnson",
  "age": 30,
  "email": "alice.johnson@example.com"
}

I'm confident this extraction is accurate based on the source material.
Please let me know if you need any modifications to the output format.`

	extractor := NewExtractor(&ParseOptions{})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		candidates, err := extractor.ExtractJSON(input)
		if err != nil {
			b.Fatal(err)
		}
		if len(candidates) == 0 {
			b.Fatal("no candidates found")
		}
	}
}
