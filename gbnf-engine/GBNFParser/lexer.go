package GBNFParser

import (
	"fmt"
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
	TokenSubExpression
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

type LexerError struct {
	Message string
	Line    int
	Column  int
	Length  int
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
		line:   0,
		column: 0,
		state:  StateDefault,
	}
}

func (lexer *Lexer) LexAllTokens() ([]*Token, []*LexerError) {
	tokens := []*Token{}
	var errors []*LexerError
	previousPos := 0
	for lexer.pos < len(lexer.input) {
		newToken, err := lexer.NextToken()
		if lexer.pos == previousPos {
			errors = append(errors, &LexerError{Message: "Lexer entered a loop."})
			break
		}
		if err != nil {
			errors = append(errors, err)
		}
		if newToken != nil {
			tokens = append(tokens, newToken)
		}
	}
	return tokens, errors
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
		lexer.column = 0
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
}

func (lexer *Lexer) NextToken() (*Token, *LexerError) {
	for lexer.pos < len(lexer.input) {
		lexer.skipWhitespace()

		startLine, startColumn := lexer.line, lexer.column
		char := lexer.peek()

		switch {
		case char == 0:
			lexer.next()
			return &Token{Type: TokenEOL, Line: startLine, Column: startColumn}, nil
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
			return &Token{Type: TokenAlternative, Value: string(char), Line: startLine, Column: startColumn}, nil
		case strings.Contains("()", string(char)):
			char = lexer.next()
			return &Token{Type: TokenSubExpression, Value: string(char), Line: startLine, Column: startColumn}, nil
		case char == ':' && lexer.pos+2 < len(lexer.input) && string(lexer.input[lexer.pos:lexer.pos+3]) == "::=":
			lexer.pos += 3
			lexer.column += 3
			return &Token{Type: TokenAssignment, Value: "::=", Line: startLine, Column: startColumn}, nil
		case strings.Contains("*?+", string(char)):
			char = lexer.next()
			return &Token{Type: TokenOperator, Value: string(char), Line: startLine, Column: startColumn}, nil
		case char == '\n':
			lexer.next()
			return &Token{Type: TokenEOL, Value: "\n", Line: startLine, Column: startColumn}, nil
		default:
			return lexer.lexUnknown()
		}

	}

	return &Token{Type: TokenEOL, Line: lexer.line, Column: lexer.column}, nil
}

func (lexer *Lexer) lexString() (*Token, *LexerError) {
	startLine, startColumn := lexer.line, lexer.column

	var value []rune
	lexer.next()
	for {
		char := lexer.next()
		if char == '"' {
			break
		}
		if char == 0 || char == '\n' {
			return nil, &LexerError{Message: fmt.Sprintf(`unterminated string`), Line: lexer.line, Column: lexer.column, Length: 1}
		}
		value = append(value, char)
	}
	return &Token{Type: TokenString, Value: string(value), Line: startLine, Column: startColumn}, nil
}

func (lexer *Lexer) lexRegex() (*Token, *LexerError) {
	startLine, startColumn := lexer.line, lexer.column
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
			return nil, &LexerError{Message: fmt.Sprintf(`unterminated regex`), Line: lexer.line, Column: lexer.column, Length: 1}
		}
		value = append(value, char)
	}
	return &Token{Type: TokenRegexp, Value: string(value), Line: startLine, Column: startColumn}, nil
}

func (lexer *Lexer) lexRange() (*Token, *LexerError) {
	startLine, startColumn := lexer.line, lexer.column
	var value []rune

	char := lexer.peek()
	for char != '}' {
		char = lexer.next()

		if char == 0 || char == '\n' {
			return nil, &LexerError{Message: fmt.Sprintf(`unterminated {} operator`), Line: lexer.line, Column: lexer.column, Length: 1}
		}
		value = append(value, char)
	}

	parts := strings.Split(string(value[1:len(value)-1]), ",")
	if len(parts) > 2 {
		return nil, &LexerError{Message: fmt.Sprintf(`unknown contents of {} operator`), Line: lexer.line, Column: lexer.column, Length: 1}
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, &LexerError{Message: fmt.Sprintf(`empty numeric value in {}`), Line: lexer.line, Column: lexer.column, Length: 1}
		}
		for _, char := range part {
			if !unicode.IsDigit(char) {
				return nil, &LexerError{Message: fmt.Sprintf(`contents of {} must be numeric`), Line: lexer.line, Column: lexer.column, Length: 1}
			}
		}
	}

	return &Token{Type: TokenRepeat, Value: string(value), Line: startLine, Column: startColumn}, nil
}

func (lexer *Lexer) lexIdentifier() (*Token, *LexerError) {
	startLine, startColumn := lexer.line, lexer.column
	breakCharacters := "\n{*+?)"
	var value []rune
	for {
		peek := lexer.peek()
		if unicode.IsLetter(peek) || unicode.IsDigit(peek) {
			value = append(value, lexer.next())
		} else if unicode.IsSpace(peek) || peek == 0 || strings.Contains(breakCharacters, string(peek)) {
			break
		} else {
			lexer.next()
			return nil, &LexerError{Message: fmt.Sprintf(`unknown variable name`), Line: lexer.line, Column: lexer.column, Length: 1}
		}
	}
	return &Token{Type: TokenIdentifier, Value: string(value), Line: startLine, Column: startColumn}, nil
}

func (lexer *Lexer) lexComment() *Token {
	char := lexer.next()
	for char != '\n' && char != 0 {
		char = lexer.next()
	}
	return &Token{Type: TokenEOL, Line: lexer.line, Column: lexer.column}
}

func (lexer *Lexer) lexUnknown() (*Token, *LexerError) {

	startLine, startColumn := lexer.line, lexer.column
	word := ""
	for !unicode.IsSpace(lexer.peek()) && lexer.pos < len(lexer.input) {
		char := lexer.next()
		word = word + string(char)
	}
	return &Token{
			Type:   TokenUnknown,
			Value:  word,
			Line:   startLine,
			Column: startColumn,
		}, &LexerError{
			Message: "unknown token",
			Line:    startLine,
			Column:  startColumn,
			Length:  len(word),
		}
}
