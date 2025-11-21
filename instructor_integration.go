package sap

import (
	"reflect"
)

// InstructorParser wraps SAP for use with instructor-go
// This allows you to use SAP as a custom parser for instructor-go
//
// Example usage:
//
//	import (
//	    "github.com/567-labs/instructor-go/pkg/instructor"
//	    "github.com/alcarpenter/gsap"
//	)
//
//	// Create a SAP-based parser
//	parser := sap.NewInstructorParser()
//
//	// Use with instructor-go
//	client := instructor.FromOpenAI(openaiClient,
//	    instructor.WithParser(parser),
//	)
type InstructorParser struct {
	parser *sapParser
}

// NewInstructorParser creates a new instructor-go compatible parser using SAP
func NewInstructorParser() *InstructorParser {
	return &InstructorParser{
		parser: NewParser(),
	}
}

// Unmarshal implements the instructor-go Parser interface
// It takes the LLM response and parses it into the target type
func (ip *InstructorParser) Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return nil
	}

	// Get the type of the target
	targetType := reflect.TypeOf(v)

	// If v is a pointer, dereference to get the actual type
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	// Parse the response
	result, err := ip.parser.Parse(string(data), targetType)
	if err != nil {
		return err
	}

	// Set the result into v
	reflect.ValueOf(v).Elem().Set(reflect.ValueOf(result))
	return nil
}

// WithStrict creates a new parser in strict mode
func (ip *InstructorParser) WithStrict(strict bool) *InstructorParser {
	ip.parser.WithStrict(strict)
	return ip
}

// WithIncompleteJSON allows incomplete JSON for streaming
func (ip *InstructorParser) WithIncompleteJSON(allow bool) *InstructorParser {
	ip.parser.WithIncompleteJSON(allow)
	return ip
}

/*
Integration Example with instructor-go:

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/567-labs/instructor-go/pkg/instructor"
	"github.com/sashabaranov/go-openai"
	"github.com/alcarpenter/gsap"
)

type User struct {
	Name  string `json:"name" jsonschema:"description=The user's full name"`
	Age   int    `json:"age" jsonschema:"description=The user's age in years"`
	Email string `json:"email" jsonschema:"description=The user's email address"`
}

func main() {
	// Create OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	openaiClient := openai.NewClient(apiKey)

	// Create SAP-based parser
	sapParser := sap.NewInstructorParser()

	// Create instructor client with SAP parser
	client := instructor.FromOpenAI(openaiClient,
		instructor.WithParser(sapParser),
	)

	// Use it to extract structured data from natural language
	ctx := context.Background()
	user := new(User)

	_, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Extract the user info: My name is Alice and I'm 30 years old. My email is alice@example.com",
				},
			},
		},
		user,
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Extracted user: %+v\n", user)
	// Output: Extracted user: &{Name:Alice Age:30 Email:alice@example.com}
}
*/
