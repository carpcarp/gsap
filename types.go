package sap

import (
	"reflect"
)

// CompletionState represents the parsing completion status
type CompletionState int

const (
	// Complete means all required fields are present
	Complete CompletionState = iota
	// Incomplete means some required fields are missing (for streaming)
	Incomplete
	// Pending means parsing is ongoing
	Pending
)

// Score represents the quality of a parse result
// Lower scores are better
type Score struct {
	flags map[string]int
	total int
}

// Flag adds to the score
func (s *Score) AddFlag(flag string, value int) {
	if s.flags == nil {
		s.flags = make(map[string]int)
	}
	s.flags[flag] = value
	s.total += value
}

// Total returns the total score
func (s *Score) Total() int {
	return s.total
}

// Less returns true if this score is better than other
func (s *Score) Less(other *Score) bool {
	return s.total < other.total
}

// ParseResult represents a successful parse
type ParseResult struct {
	Value             interface{}
	Score             *Score
	CompletionState   CompletionState
	RemainingContent  string // Text that wasn't part of JSON
}

// JSONCandidate represents a potential JSON string extracted from text
type JSONCandidate struct {
	JSON  string // The JSON string
	Index int    // Starting position in original text
}

// Parser defines the interface for extracting JSON from text
type Parser interface {
	// Parse extracts potential JSON candidates from input text
	Parse(input string) ([]JSONCandidate, error)
}

// Coercer defines the interface for type coercion
type Coercer interface {
	// Coerce transforms a value to match the target type
	Coerce(value interface{}, targetType reflect.Type) (interface{}, *Score, error)
}

// FixingParser handles malformed JSON
type FixingParser struct {
	allowIncomplete bool // Allow incomplete JSON for streaming
}

// Extractor handles JSON extraction from text
type Extractor struct {
	parser *FixingParser
}

// StreamingOptions configures streaming behavior
type StreamingOptions struct {
	AllowIncompleteJSON bool
	TrackCompletionState bool
}

// ParseOptions configures parsing behavior
type ParseOptions struct {
	Streaming StreamingOptions
	Strict    bool // If true, only accept exact JSON matches
}
