package src

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenUnknown TokenType = iota
	TokenAssignment
	TokenString
	TokenRegexp
	TokenOperator
	TokenAlternative
	TokenExpression
	TokenIdentifier
	TokenRepeat
	TokenEOL
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

type LexerState int

const (
	StateDefault LexerState = iota
	StateWord
)

type Lexer struct {
	input  []rune
	pos    int
	line   int
	column int
	state  LexerState
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  []rune(input),
		pos:    0,
		line:   1,
		column: 1,
		state:  StateDefault,
	}
}

func (lexer *Lexer) peek() rune {
	if lexer.pos >= len(lexer.input) {
		return 0
	}
	return lexer.input[lexer.pos]
}

func (lexer *Lexer) next() rune {
	char := lexer.peek()
	if char == 0 {
		return 0
	}

	lexer.pos++
	if char == '\n' {
		lexer.line++
		lexer.column = 1
	} else {
		lexer.column++
	}

	return char
}

func (lexer *Lexer) skipWhitespace() {
	nSkipped := 0
	for unicode.IsSpace(lexer.peek()) && lexer.peek() != '\n' {
		lexer.next()
		nSkipped++
	}

	if nSkipped == 0 && lexer.column != 1 {
		lexer.state = StateWord
	} else {
		lexer.state = StateDefault
	}
}

func (lexer *Lexer) NextToken() (Token, error) {
	for lexer.pos < len(lexer.input) {
		lexer.skipWhitespace()

		startLine, startColumn := lexer.line, lexer.column
		char := lexer.peek()

		switch {
		case char == 0:
			return Token{Type: TokenEOL, Line: startLine, Column: startColumn}, io.EOF
		case char == '#':
			return lexer.lexComment(), nil
		case char == '"':
			return lexer.lexString()
		case char == '[':
			return lexer.lexRegex()
		case char == '{':
			return lexer.lexRange()
		case unicode.IsLetter(char):
			return lexer.lexIdentifier()
		case char == '|':
			char = lexer.next()
			return Token{Type: TokenAlternative, Value: string(char), Line: startLine, Column: startColumn}, nil
		case strings.Contains("()", string(char)):
			char = lexer.next()
			return Token{Type: TokenExpression, Value: string(char), Line: startLine, Column: startColumn}, nil
		case char == ':' && lexer.pos+2 < len(lexer.input) && string(lexer.input[lexer.pos:lexer.pos+3]) == "::=":
			lexer.pos += 3
			lexer.column += 3
			return Token{Type: TokenAssignment, Value: "::=", Line: startLine, Column: startColumn}, nil
		case strings.Contains("*?+", string(char)):
			char = lexer.next()
			return Token{Type: TokenOperator, Value: string(char), Line: startLine, Column: startColumn}, nil
		case char == '\n':
			lexer.next()
			return Token{Type: TokenEOL, Value: "\n", Line: startLine, Column: startColumn}, nil
		default:
			return Token{}, fmt.Errorf("unexpected character '%c' at %d:%d", char, lexer.line, lexer.column)
		}

	}

	return Token{Type: TokenEOL, Line: lexer.line, Column: lexer.column}, io.EOF
}

func (lexer *Lexer) lexString() (Token, error) {
	startLine, startColumn := lexer.line, lexer.column
	if lexer.state != StateDefault {
		return Token{}, fmt.Errorf(`unexpected character " at %d:%d`, startLine, startColumn)
	}

	var value []rune
	lexer.next()
	for {
		char := lexer.next()
		if char == '"' {
			break
		}
		if char == 0 || char == '\n' {
			return Token{}, fmt.Errorf("unterminated string at %d:%d", startLine, startColumn)
		}
		value = append(value, char)
	}
	return Token{Type: TokenString, Value: string(value), Line: startLine, Column: startColumn}, nil
}

func (lexer *Lexer) lexRegex() (Token, error) {
	startLine, startColumn := lexer.line, lexer.column
	if lexer.state != StateDefault {
		return Token{}, fmt.Errorf(`unexpected character [ at %d:%d`, startLine, startColumn)
	}
	var value []rune
	bracketCount := 0
	for {
		char := lexer.next()
		if char == '[' {
			bracketCount++
		}
		if char == ']' {
			bracketCount--
			if bracketCount == 0 {
				value = append(value, char)
				break
			}
		}
		if char == 0 || char == '\n' {
			return Token{}, fmt.Errorf("unterminated regex at %d:%d", startLine, startColumn)
		}
		value = append(value, char)
	}
	return Token{Type: TokenRegexp, Value: string(value), Line: startLine, Column: startColumn}, nil
}

func (lexer *Lexer) lexRange() (Token, error) {
	startLine, startColumn := lexer.line, lexer.column
	if lexer.state != StateWord {
		return Token{}, fmt.Errorf(`unexpected character { at %d:%d`, startLine, startColumn)
	}
	var value []rune

	char := lexer.peek()
	for char != '}' {
		char = lexer.next()

		if char == 0 || char == '\n' {
			return Token{}, fmt.Errorf("unterminated {} operator at %d:%d", startLine, startColumn)
		}
		value = append(value, char)
	}

	parts := strings.Split(string(value[1:len(value)-1]), ",")
	if len(parts) > 2 {
		return Token{}, fmt.Errorf("unknown contents of {} operator at %d:%d", startLine, startColumn)
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return Token{}, fmt.Errorf("empty numeric value in {} at %d:%d", startLine, startColumn)
		}
		for _, char := range part {
			if !unicode.IsDigit(char) {
				return Token{}, fmt.Errorf("contents of {} must be numeric at %d:%d", startLine, startColumn)
			}
		}
	}

	return Token{Type: TokenRepeat, Value: string(value), Line: startLine, Column: startColumn}, nil
}

func (lexer *Lexer) lexIdentifier() (Token, error) {
	startLine, startColumn := lexer.line, lexer.column
	if lexer.state != StateDefault {
		return Token{}, fmt.Errorf(`unexpected character %s at %d:%d`, string(lexer.peek()), startLine, startColumn)
	}
	breakCharacters := "\n{*+?"
	var value []rune
	for {
		peek := lexer.peek()
		if unicode.IsLetter(peek) || unicode.IsDigit(peek) {
			value = append(value, lexer.next())
		} else if unicode.IsSpace(peek) || peek == 0 || strings.Contains(breakCharacters, string(peek)) {
			break
		} else {
			return Token{}, fmt.Errorf("unknown variable name at %d:%d", startLine, startColumn)
		}
	}
	return Token{Type: TokenIdentifier, Value: string(value), Line: startLine, Column: startColumn}, nil
}

func (lexer *Lexer) lexComment() Token {
	char := lexer.next()
	for char != '\n' && char != 0 {
		char = lexer.next()
	}
	return Token{Type: TokenEOL, Line: lexer.line, Column: lexer.column}
}
