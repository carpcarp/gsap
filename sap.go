package sap

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// DefaultParser is the default SAP parser instance
var DefaultParser = NewParser()

// NewParser creates a new SAP parser with default options
func NewParser() *sapParser {
	return &sapParser{
		options: &ParseOptions{
			Streaming: StreamingOptions{
				AllowIncompleteJSON: false,
				TrackCompletionState: true,
			},
			Strict: false,
		},
	}
}

// sapParser is the main parser implementation
type sapParser struct {
	options *ParseOptions
	extractor *Extractor
	coercer *TypeCoercer
}

// Parse parses input text into the target type
// This is the main public API
func Parse[T any](input string) (T, error) {
	var zero T
	result, err := DefaultParser.Parse(input, reflect.TypeOf(zero))
	if err != nil {
		return zero, err
	}
	typed, ok := result.(T)
	if !ok {
		return zero, fmt.Errorf("type mismatch: expected %T, got %T", zero, result)
	}
	return typed, nil
}

// ParseWithScore is like Parse but also returns the parse score
func ParseWithScore[T any](input string) (T, *Score, error) {
	var zero T
	result, score, err := DefaultParser.ParseWithScore(input, reflect.TypeOf(zero))
	if err != nil {
		return zero, nil, err
	}
	typed, ok := result.(T)
	if !ok {
		return zero, nil, fmt.Errorf("type mismatch: expected %T, got %T", zero, result)
	}
	return typed, score, nil
}

// ParsePartial parses input as a partial type (for streaming)
func ParsePartial[T any](input string) (T, CompletionState, error) {
	var zero T
	result, state, err := DefaultParser.ParsePartial(input, reflect.TypeOf(zero))
	if err != nil {
		return zero, Complete, err
	}
	typed, ok := result.(T)
	if !ok {
		return zero, state, fmt.Errorf("type mismatch: expected %T, got %T", zero, result)
	}
	return typed, state, nil
}

// Parse implements the Parser interface
func (p *sapParser) Parse(input string, targetType reflect.Type) (interface{}, error) {
	result, _, err := p.ParseWithScore(input, targetType)
	return result, err
}

// ParseWithScore extracts and parses JSON, returning the best match
func (p *sapParser) ParseWithScore(input string, targetType reflect.Type) (interface{}, *Score, error) {
	if p.extractor == nil {
		p.extractor = NewExtractor(&ParseOptions{
			Streaming: p.options.Streaming,
			Strict:    p.options.Strict,
		})
	}
	if p.coercer == nil {
		p.coercer = NewTypeCoercer()
	}

	// Extract potential JSON candidates
	candidates, err := p.extractor.ExtractJSON(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract JSON: %w", err)
	}

	if len(candidates) == 0 {
		return nil, nil, fmt.Errorf("no JSON found in input")
	}

	// Try to parse and coerce each candidate, pick the best
	var bestResult interface{}
	var bestScore *Score
	var bestErr error

	for _, candidate := range candidates {
		// Unmarshal raw JSON
		var rawValue interface{}
		if err := json.Unmarshal([]byte(candidate.JSON), &rawValue); err != nil {
			// If strict mode, skip on parse errors
			if p.options.Strict {
				continue
			}
			// Otherwise, try to fix it
			fixed, fixErr := FixJSON(candidate.JSON)
			if fixErr != nil {
				bestErr = fixErr
				continue
			}
			if err := json.Unmarshal([]byte(fixed), &rawValue); err != nil {
				bestErr = err
				continue
			}
		}

		// Coerce to target type
		result, candScore, err := p.coercer.Coerce(rawValue, targetType)
		if err != nil {
			bestErr = err
			continue
		}

		// Keep the best result
		if bestScore == nil || candScore.Less(bestScore) {
			bestResult = result
			bestScore = candScore
			bestErr = nil
		}
	}

	if bestErr != nil || bestScore == nil {
		return nil, nil, fmt.Errorf("failed to parse: %w", bestErr)
	}

	return bestResult, bestScore, nil
}

// ParsePartial parses as a partial type (streaming)
func (p *sapParser) ParsePartial(input string, targetType reflect.Type) (interface{}, CompletionState, error) {
	result, _, err := p.ParseWithScore(input, targetType)
	if err != nil {
		return nil, Complete, err
	}

	// Determine completion state based on required fields
	state := Complete
	if p.options.Streaming.TrackCompletionState {
		// TODO: Track which required fields are missing
		state = Complete
	}

	return result, state, nil
}

// WithStrict creates a new parser in strict mode (no fixing)
func (p *sapParser) WithStrict(strict bool) *sapParser {
	p.options.Strict = strict
	return p
}

// WithIncompleteJSON allows incomplete JSON for streaming
func (p *sapParser) WithIncompleteJSON(allow bool) *sapParser {
	p.options.Streaming.AllowIncompleteJSON = allow
	return p
}
