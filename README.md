# GSAP - Go Schema-Aligned Parser

A robust JSON parser for Go inspired by BAML's Schema-Aligned Parsing (SAP) algorithm. GSAP handles messy LLM-generated JSON by gracefully recovering from common issues like missing quotes, trailing commas, single quotes, and type mismatches.

## Features

✅ **Robust JSON Extraction**
- Extract JSON from markdown code blocks
- Find JSON embedded in natural language
- Handle chain-of-thought reasoning before/after JSON

✅ **Intelligent JSON Fixing**
- Fix unquoted keys and values
- Convert single/triple quotes to double quotes
- Remove trailing commas
- Auto-close unclosed structures
- Skip comments (`//`, `/* */`)

✅ **Smart Type Coercion**
- String → Int/Float/Bool with intelligent parsing
- Fraction parsing (`"1/5"` → 0.2)
- Currency parsing (`"$1,234.56"` → 1234.56)
- Comma-separated number parsing
- Array and struct field coercion
- Case-insensitive enum matching

✅ **Fuzzy String Matching**
- Unicode accent removal (`étude` → `etude`)
- Ligature expansion (`æ` → `ae`)
- Levenshtein distance-based matching
- Enum value disambiguation

✅ **Type-Safe Parsing**
- Works with standard Go structs
- Automatic field matching (JSON tags, case-insensitive fallback)
- Score-based best match selection

## Installation

```bash
go get github.com/alcarpenter/gsap
```

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/alcarpenter/gsap"
)

type User struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func main() {
	// Even with messy input, GSAP parses it correctly
	input := `{name: "Alice", age: "30", email: "alice@example.com"}`

	user, err := gsap.Parse[User](input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s is %d years old\n", user.Name, user.Age)
	// Output: Alice is 30 years old
}
```

## Usage Examples

### Basic Parsing

```go
input := `{"name": "Bob", "age": 25}`
user, err := gsap.Parse[User](input)
```

### JSON in Markdown Code Blocks

```go
input := `Here's the user data:
\`\`\`json
{
  "name": "Charlie",
  "age": 35,
  "email": "charlie@example.com"
}
\`\`\``

user, err := gsap.Parse[User](input)
```

### Unquoted Keys and Values

```go
input := `{name: Bob, age: 28, email: bob@example.com}`
user, err := gsap.Parse[User](input)
```

### Type Coercion

```go
input := `{"name": "David", "age": "40", "email": "david@example.com"}`
user, err := gsap.Parse[User](input)
// age is a string in JSON but parsed as int
```

### With Score (Best Match Quality)

```go
user, score, err := gsap.ParseWithScore[User](input)
if score != nil {
	fmt.Printf("Parse score: %d\n", score.Total())
}
```

### Struct with Complex Types

```go
type Resume struct {
	Title      string   `json:"title"`
	Experience []string `json:"experience"`
	Active     bool     `json:"active"`
}

input := `{"title": "Engineer", "experience": ["Go", "Rust"], "active": "yes"}`
resume, err := gsap.Parse[Resume](input)
```

### Manual JSON Fixing

```go
messy := `{name: John, age: 30,}`  // Unquoted, trailing comma
fixed, err := gsap.FixJSON(messy)
// fixed: `{"name": "John", "age": 30}`
```

## How It Works

### 1. JSON Extraction
SAP first extracts potential JSON from the input text:
1. Try standard JSON parsing
2. Look for markdown code blocks (` ```json ... ``` `)
3. Find balanced JSON objects/arrays in text
4. Fall back to fixing malformed JSON

### 2. JSON Fixing
When JSON is malformed, SAP's fixing parser:
- Tracks open/close brackets and quotes
- Automatically quotes unquoted keys/values
- Converts single/triple quotes to double quotes
- Removes comments and trailing commas
- Auto-closes incomplete structures

### 3. Type Coercion
After getting valid JSON, SAP coerces values to match the target type:
- String "42" → int 42
- String "true" → bool true
- Float 3.7 → int 4 (rounds)
- "1/5" → float 0.2
- Fuzzy matches enum values

### 4. Best Match Selection
When multiple parsings are valid, SAP picks the best using a scoring system:
- Exact matches score lowest (best)
- Type coercions score higher
- Fuzzy matches score highest (worst)

## Configuration

### Strict Mode (No Fixing)

```go
parser := gsap.NewParser().WithStrict(true)
result, _, _ := parser.ParseWithScore(input, reflect.TypeOf(User{}))
```

### Incomplete JSON for Streaming

```go
parser := gsap.NewParser().WithIncompleteJSON(true)
// Allows parsing of incomplete JSON for streaming responses
```

## Integration with instructor-go

SAP can be integrated as a custom parser for [instructor-go](https://github.com/567-labs/instructor-go):

```go
import (
	"github.com/567-labs/instructor-go/pkg/instructor"
	"github.com/alcarpenter/sap"
)

// Create a custom parser
type sapParser struct{}

func (p *sapParser) Unmarshal(data []byte, v any) error {
	result, err := gsap.Parse[T](string(data))
	// Type assertion and assignment...
	return err
}

// Use with instructor-go
client := instructor.FromOpenAI(openaiClient)
// Integration details...
```

## Performance

- **Extraction**: O(n) single pass through input text
- **Fixing**: O(n) state machine
- **Coercion**: O(m) where m is number of struct fields
- **Overall**: Near-linear performance for typical LLM outputs

## Comparison to BAML

| Feature | SAP | BAML |
|---------|-----|------|
| JSON fixing | ✅ | ✅ |
| Type coercion | ✅ | ✅ |
| Fuzzy matching | ✅ | ✅ |
| Language | Go | Rust (generates Go) |
| Schema DSL | ❌ | ✅ |
| Streaming support | 🔄 | ✅ |
| Multi-provider | ❌ | ✅ |
| Type safety | ✅ | ✅ |

## Limitations

- No discriminated unions (use `interface{}` with type switching)
- No streaming partial types yet (Incomplete/Pending states)
- No constraint validation (`@check`, `@assert`)
- No expression evaluation
- Go's static type system limits dynamic types

## Testing

```bash
go test -v
```

All tests pass, including:
- Valid JSON parsing
- JSON in markdown code blocks
- Unquoted keys and values
- String to type coercion
- Trailing comma handling
- Boolean coercion
- Array parsing
- Fuzzy string matching

## Roadmap

- [ ] Streaming support with partial types
- [ ] Constraint validation
- [ ] Union type support
- [ ] Schema generation from structs
- [ ] Field description parsing
- [ ] Custom coercer registration
- [ ] Performance benchmarks

## Contributing

PRs welcome! Focus areas:
- Streaming/partial type support
- Union type handling
- Constraint validation
- Performance optimizations

## License

MIT

## Acknowledgments

Inspired by [BAML](https://www.boundaryml.com)'s Schema-Aligned Parsing algorithm. SAP extracts this powerful parsing capability into a lightweight, pure-Go library for use with any Go LLM client.
