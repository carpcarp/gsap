package sap

import (
	"strings"
	"unicode"
)

// FixJSON attempts to fix malformed JSON
func FixJSON(input string) (string, error) {
	parser := &fixingParserState{
		input: input,
		runes: []rune(input),
	}
	return parser.parse()
}

type fixingParserState struct {
	input  string
	runes  []rune
	pos    int
	result strings.Builder

	// State tracking
	inString           bool
	stringQuoteChar    rune
	stringEscaped      bool
	lastNonWhitespace  rune
	bracketStack       []rune // Stack of open brackets/braces
	unquotedValueStart int    // Position where unquoted value started
}

func (p *fixingParserState) parse() (string, error) {
	for p.pos < len(p.runes) {
		ch := p.runes[p.pos]

		if p.inString {
			p.handleStringChar(ch)
		} else {
			p.handleNonStringChar(ch)
		}

		p.pos++
	}

	// Close any unclosed structures
	p.closeUnclosedStructures()

	return p.result.String(), nil
}

func (p *fixingParserState) handleStringChar(ch rune) {
	if p.stringEscaped {
		p.result.WriteRune(ch)
		p.stringEscaped = false
		return
	}

	if ch == '\\' {
		p.result.WriteRune(ch)
		p.stringEscaped = true
		return
	}

	if ch == p.stringQuoteChar {
		// End of string
		p.result.WriteRune('"')
		p.inString = false
		p.lastNonWhitespace = '"'
		return
	}

	p.result.WriteRune(ch)
}

func (p *fixingParserState) handleNonStringChar(ch rune) {
	switch ch {
	case '"':
		// Start of quoted string
		p.result.WriteRune('"')
		p.inString = true
		p.stringQuoteChar = '"'
		p.lastNonWhitespace = '"'

	case '\'':
		// Single quote - convert to double quote
		p.result.WriteRune('"')
		p.inString = true
		p.stringQuoteChar = '\''
		p.lastNonWhitespace = '"'

	case '`':
		// Backtick - convert to double quote
		p.result.WriteRune('"')
		p.inString = true
		p.stringQuoteChar = '`'
		p.lastNonWhitespace = '"'

	case '{', '[':
		p.handleOpenBracket(ch)

	case '}', ']':
		p.handleCloseBracket(ch)

	case ':':
		// Quote any unquoted key before the colon
		p.quoteUnquotedKey()
		p.result.WriteRune(ch)
		p.lastNonWhitespace = ch

	case ',':
		// Quote any unquoted value before the comma
		if p.lastNonWhitespace != '{' && p.lastNonWhitespace != '[' && p.lastNonWhitespace != ',' && p.lastNonWhitespace != '"' {
			p.quoteUnquotedValue()
		}
		p.result.WriteRune(ch)
		p.lastNonWhitespace = ch

	case ' ', '\t', '\n', '\r':
		// Whitespace - collapse outside strings
		str := p.result.String()
		if len(str) > 0 && !strings.HasSuffix(strings.TrimRight(str, " \t\n\r"), " ") {
			if p.lastNonWhitespace != '{' && p.lastNonWhitespace != '[' && p.lastNonWhitespace != ',' && p.lastNonWhitespace != ':' && !unicode.IsSpace(p.lastNonWhitespace) {
				p.result.WriteRune(' ')
			}
		}

	case '/', '*':
		// Comments - try to skip
		if ch == '/' && p.pos+1 < len(p.runes) && p.runes[p.pos+1] == '/' {
			// Line comment
			p.pos++
			for p.pos < len(p.runes) && p.runes[p.pos] != '\n' {
				p.pos++
			}
		} else if ch == '/' && p.pos+1 < len(p.runes) && p.runes[p.pos+1] == '*' {
			// Block comment
			p.pos++
			for p.pos+1 < len(p.runes) {
				if p.runes[p.pos] == '*' && p.runes[p.pos+1] == '/' {
					p.pos++
					break
				}
				p.pos++
			}
		}

	default:
		// Regular character - might be part of unquoted key or value
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-' || ch == '+' || ch == '.' || ch == '_' {
			p.result.WriteRune(ch)
			if !unicode.IsSpace(ch) {
				p.lastNonWhitespace = ch
			}
		} else if ch == 'n' || ch == 't' || ch == 'f' || ch == 'e' || ch == 'E' {
			// Might be null, true, false, or exponent
			p.result.WriteRune(ch)
			p.lastNonWhitespace = ch
		}
	}
}

func (p *fixingParserState) quoteUnquotedKey() {
	// Find and quote any unquoted key
	str := p.result.String()
	str = strings.TrimRight(str, " \t\n\r")

	// Find the start of unquoted text
	lastStructuralChar := -1
	for i := len(str) - 1; i >= 0; i-- {
		ch := str[i]
		if ch == '{' || ch == ',' {
			lastStructuralChar = i
			break
		}
	}

	if lastStructuralChar >= 0 && lastStructuralChar < len(str)-1 {
		// There's unquoted text after the last structural character
		before := str[:lastStructuralChar+1]
		unquoted := strings.TrimSpace(str[lastStructuralChar+1:])

		if len(unquoted) > 0 && !startsWithQuote(unquoted) {
			// Quote the unquoted key
			p.result.Reset()
			p.result.WriteString(before)
			p.result.WriteString(" \"")
			p.result.WriteString(unquoted)
			p.result.WriteString("\"")
		}
	}
}

func (p *fixingParserState) quoteUnquotedValue() {
	// Find and quote any unquoted value
	str := p.result.String()
	str = strings.TrimRight(str, " \t\n\r")

	// Find the last structural character
	lastStructuralChar := -1
	for i := len(str) - 1; i >= 0; i-- {
		ch := str[i]
		if ch == ':' || ch == '[' || ch == ',' {
			lastStructuralChar = i
			break
		}
	}

	if lastStructuralChar >= 0 && lastStructuralChar < len(str)-1 {
		// There's unquoted text after the last structural character
		before := str[:lastStructuralChar+1]
		unquoted := strings.TrimSpace(str[lastStructuralChar+1:])

		if len(unquoted) > 0 && !startsWithQuote(unquoted) && !isReservedWord(unquoted) {
			// Quote the unquoted value
			p.result.Reset()
			p.result.WriteString(before)
			p.result.WriteString(" \"")
			p.result.WriteString(unquoted)
			p.result.WriteString("\"")
		}
	}
}

func (p *fixingParserState) closeUnclosedStructures() {
	// Close any remaining open brackets in reverse order
	for len(p.bracketStack) > 0 {
		lastOpen := p.bracketStack[len(p.bracketStack)-1]
		p.bracketStack = p.bracketStack[:len(p.bracketStack)-1]

		// Quote any pending unquoted value
		if lastOpen == '}' {
			p.quoteUnquotedValue()
		}

		if lastOpen == '{' {
			p.removeTrailingComma()
			p.result.WriteRune('}')
		} else if lastOpen == '[' {
			p.removeTrailingComma()
			p.result.WriteRune(']')
		}
	}
}

func (p *fixingParserState) removeTrailingComma() {
	// Remove trailing comma if present
	str := p.result.String()
	str = strings.TrimRight(str, " \t\n\r")
	if strings.HasSuffix(str, ",") {
		str = str[:len(str)-1]
		p.result.Reset()
		p.result.WriteString(str)
	}
}

func (p *fixingParserState) handleOpenBracket(ch rune) {
	p.result.WriteRune(ch)
	p.bracketStack = append(p.bracketStack, ch)
	p.lastNonWhitespace = ch
}

func (p *fixingParserState) handleCloseBracket(ch rune) {
	// Check if we have an open bracket to match
	if len(p.bracketStack) == 0 {
		return
	}

	lastOpen := p.bracketStack[len(p.bracketStack)-1]
	expectedClose := rune(0)

	if lastOpen == '{' {
		expectedClose = '}'
	} else if lastOpen == '[' {
		expectedClose = ']'
	}

	if ch == expectedClose {
		// Quote any pending unquoted value
		if expectedClose == '}' {
			p.quoteUnquotedValue()
		}

		p.bracketStack = p.bracketStack[:len(p.bracketStack)-1]
		// Remove trailing comma before closing bracket
		p.removeTrailingComma()
		p.result.WriteRune(ch)
		p.lastNonWhitespace = ch
	}
}

func startsWithQuote(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] == '"' || s[0] == '\'' || s[0] == '`'
}

func isReservedWord(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch lower {
	case "true", "false", "null":
		return true
	default:
		// Check if it looks like a number
		_, err := parseNumber(s)
		return err == nil
	}
}
