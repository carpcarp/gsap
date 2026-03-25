package sap

import (
	"fmt"
	"reflect"
	"time"
)

type TestCandidate struct {
	Name       string    `json:"name"`
	Age        int       `json:"age"`
	Email      string    `json:"email"`
	Skills     []string  `json:"skills"`
	Experience int       `json:"experience"`
	Salary     float64   `json:"salary"`
	Remote     bool      `json:"remote"`
	StartDate  time.Time `json:"start_date"`
	Website    *string   `json:"website"`
	Notes      *string   `json:"notes"`
}

func ExampleParse() {
	input := `{"name": "Alice", "age": 30, "email": "alice@example.com"}`

	user, err := Parse[TestUser](input)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	fmt.Println(user.Email)
	// Output:
	// Alice
	// 30
	// alice@example.com
}

func ExampleParse_markdown() {
	input := "Here's the user data:\n```json\n{\"name\": \"Bob\", \"age\": 25, \"email\": \"bob@example.com\"}\n```"

	user, err := Parse[TestUser](input)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	// Output:
	// Bob
	// 25
}

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

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	fmt.Println(user.Email)
	// Output:
	// Alice
	// 30
	// alice@example.com
}

func ExampleParse_typeCoercion() {
	// SAP coerces string "30" to int 30 automatically
	input := `{"name": "Alice", "age": "30", "email": "alice@example.com"}`

	user, err := Parse[TestUser](input)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	// Output:
	// Alice
	// 30
}

func ExampleParse_booleanCoercion() {
	// SAP coerces string "yes" to bool true
	input := `{"title": "Dev", "experience": ["Go"], "active": "yes"}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		panic(err)
	}

	fmt.Println(resume.Title)
	fmt.Println(resume.Active)
	// Output:
	// Dev
	// true
}

func ExampleParse_complexTypes() {
	input := `{"title": "Engineer", "experience": ["Go", "Rust", "Python"], "active": true}`

	resume, err := Parse[TestResume](input)
	if err != nil {
		panic(err)
	}

	fmt.Println(resume.Title)
	fmt.Println(len(resume.Experience))
	fmt.Println(resume.Experience[0])
	fmt.Println(resume.Active)
	// Output:
	// Engineer
	// 3
	// Go
	// true
}

func ExampleParseWithScore() {
	// ParseWithScore returns a quality score alongside the result.
	// Lower scores indicate cleaner input (fewer fixes needed).
	input := `{"name": "Frank", "age": "50", "email": "frank@example.com"}`

	user, score, err := ParseWithScore[TestUser](input)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	fmt.Printf("score >= 0: %v\n", score.Total() >= 0)
	// Output:
	// Frank
	// 50
	// score >= 0: true
}

func ExampleFixJSON() {
	input := `{name: "Alice", age: 30, email: "alice@example.com",}`

	fixed, err := FixJSON(input)
	if err != nil {
		panic(err)
	}

	// The fixed JSON can now be parsed by encoding/json
	user, err := Parse[TestUser](fixed)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	// Output:
	// Alice
	// 30
}

func ExampleFixJSON_singleQuotes() {
	input := `{'name': 'David', 'age': 40, 'email': 'david@example.com'}`

	fixed, err := FixJSON(input)
	if err != nil {
		panic(err)
	}

	user, err := Parse[TestUser](fixed)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	// Output:
	// David
	// 40
}

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
	user, err := Parse[TestUser](fixed)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	// Output:
	// Grace
	// 31
}

func ExampleNewParser() {
	// Create a custom parser with strict mode enabled.
	// Strict mode rejects malformed JSON instead of fixing it.
	userType := reflect.TypeOf(TestUser{})

	// Valid JSON works fine in strict mode
	parser := NewParser().WithStrict(true)
	validInput := `{"name": "Alice", "age": 30, "email": "alice@example.com"}`
	_, err := parser.Parse(validInput, userType)

	// Unquoted keys are rejected in strict mode
	invalidInput := `{name: "Alice", age: 30}`
	_, err2 := NewParser().WithStrict(true).Parse(invalidInput, userType)

	fmt.Printf("valid input error: %v\n", err)
	fmt.Printf("invalid input rejected: %v\n", err2 != nil)
	// Output:
	// valid input error: <nil>
	// invalid input rejected: true
}

func ExampleNewInstructorParser() {
	// NewInstructorParser creates a parser compatible with instructor-go.
	// Use it as a drop-in custom parser for instructor-go clients.
	parser := NewInstructorParser()

	// The parser implements Unmarshal([]byte, interface{}) error
	var user TestUser
	err := parser.Unmarshal(
		[]byte(`{"name": "Alice", "age": "30", "email": "alice@example.com"}`),
		&user,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	// Output:
	// Alice
	// 30
}

func ExampleParse_realWorldLLMOutput() {
	// This is what LLM output actually looks like — chain-of-thought preamble,
	// markdown code block, broken JSON with comments, mixed quotes, and values
	// that need coercion. GSAP handles all of it in a single call.
	input := "Sure! Here's the candidate information I extracted:\n\n" +
		"```json\n" +
		"{\n" +
		"    // Personal details\n" +
		"    name: 'Jane Doe',\n" +
		"    'age': \"29 years\",\n" +
		"    email: 'jane.doe@example.com',\n" +
		"\n" +
		"    /* Professional info */\n" +
		"    skills: \"go, python, react, typescript\",\n" +
		"    experience: \"8 years\",\n" +
		"    salary: \"$185K\",\n" +
		"    remote: \"yes\",\n" +
		"    start_date: \"2025-03-15\",\n" +
		"\n" +
		"    // Optional fields\n" +
		"    website: \"N/A\",\n" +
		"    notes: \"TBD\",\n" +
		"}\n" +
		"```\n\n" +
		"Let me know if you need anything else!"

	candidate, err := Parse[TestCandidate](input)
	if err != nil {
		panic(err)
	}

	fmt.Println(candidate.Name)
	fmt.Println(candidate.Age)
	fmt.Println(candidate.Email)
	fmt.Println(candidate.Skills)
	fmt.Println(candidate.Experience)
	fmt.Println(candidate.Salary)
	fmt.Println(candidate.Remote)
	fmt.Println(candidate.StartDate.Format("2006-01-02"))
	fmt.Println(candidate.Website)
	fmt.Println(candidate.Notes)
	// Output:
	// Jane Doe
	// 29
	// jane.doe@example.com
	// [go python react typescript]
	// 8
	// 185000
	// true
	// 2025-03-15
	// <nil>
	// <nil>
}
