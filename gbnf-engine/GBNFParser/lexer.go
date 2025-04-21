package GBNFParser

import (
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
	Error  string
}

type Lexer struct {
	input  []rune
	pos    int
	line   int
	column int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  []rune(input),
		pos:    0,
		line:   0,
		column: 0,
	}
}

func (lexer *Lexer) LexAllTokens() []Token {
	tokens := []Token{}
	previousPos := 0
	for lexer.pos < len(lexer.input) {
		newToken := lexer.nextToken()
		if lexer.pos == previousPos {
			loopToken := Token{
				Error: "lexer entered a loop",
				Type:  TokenUnknown,
			}
			tokens = append(tokens, loopToken)
			break
		}
		tokens = append(tokens, newToken)
	}
	return tokens
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

func (lexer *Lexer) nextToken() Token {
	for lexer.pos < len(lexer.input) {
		lexer.skipWhitespace()

		startLine, startColumn := lexer.line, lexer.column
		char := lexer.peek()

		switch {
		case char == 0:
			lexer.next()
			return Token{Type: TokenEOL, Line: startLine, Column: startColumn}
		case char == '#':
			return lexer.lexComment()
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
			return Token{Type: TokenAlternative, Value: string(char), Line: startLine, Column: startColumn}
		case strings.Contains("()", string(char)):
			char = lexer.next()
			return Token{Type: TokenSubExpression, Value: string(char), Line: startLine, Column: startColumn}
		case char == ':' && lexer.pos+2 < len(lexer.input) && string(lexer.input[lexer.pos:lexer.pos+3]) == "::=":
			lexer.pos += 3
			lexer.column += 3
			return Token{Type: TokenAssignment, Value: "::=", Line: startLine, Column: startColumn}
		case strings.Contains("*?+", string(char)):
			char = lexer.next()
			return Token{Type: TokenOperator, Value: string(char), Line: startLine, Column: startColumn}
		case char == '\n':
			lexer.next()
			return Token{Type: TokenEOL, Value: "\n", Line: startLine, Column: startColumn}
		default:
			return lexer.lexUnknown()
		}

	}

	return Token{Type: TokenEOL, Line: lexer.line, Column: lexer.column}
}

func (lexer *Lexer) lexString() Token {
	startLine, startColumn := lexer.line, lexer.column

	var value []rune
	lexer.next()
	for {
		char := lexer.next()
		if char == '"' {
			break
		}
		if char == 0 || char == '\n' {
			return Token{Type: TokenString, Value: string(value), Error: "unterminated string", Line: startLine, Column: startColumn}
		}
		value = append(value, char)
	}
	return Token{Type: TokenString, Value: string(value), Line: startLine, Column: startColumn}
}

func (lexer *Lexer) lexRegex() Token {
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
			return Token{Type: TokenRegexp, Value: string(value), Line: startLine, Column: startColumn, Error: "unterminated regex"}
		}
		value = append(value, char)
	}
	return Token{Type: TokenRegexp, Value: string(value), Line: startLine, Column: startColumn}
}

func (lexer *Lexer) lexRange() Token {
	startLine, startColumn := lexer.line, lexer.column
	var value []rune

	char := lexer.peek()
	for char != '}' {
		char = lexer.next()

		if char == 0 || char == '\n' {
			return Token{Type: TokenRepeat, Value: string(value), Line: startLine, Column: startColumn, Error: "unterminated {} operator"}
		}
		value = append(value, char)
	}

	parts := strings.Split(string(value[1:len(value)-1]), ",")
	if len(parts) > 2 {
		return Token{Type: TokenRepeat, Value: string(value), Line: startLine, Column: startColumn, Error: "unknown contents of {} operator"}
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return Token{Type: TokenRepeat, Value: string(value), Line: startLine, Column: startColumn, Error: "empty numeric value in {}"}
		}
		for _, char := range part {
			if !unicode.IsDigit(char) {
				return Token{Type: TokenRepeat, Value: string(value), Line: startLine, Column: startColumn, Error: "contents of {} must be numeric"}
			}
		}
	}

	return Token{Type: TokenRepeat, Value: string(value), Line: startLine, Column: startColumn}
}

func (lexer *Lexer) lexIdentifier() Token {
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
			return Token{Type: TokenIdentifier, Value: string(value), Line: startLine, Column: startColumn, Error: "unknown variable name"}
		}
	}
	return Token{Type: TokenIdentifier, Value: string(value), Line: startLine, Column: startColumn}
}

func (lexer *Lexer) lexComment() Token {
	char := lexer.next()
	for char != '\n' && char != 0 {
		char = lexer.next()
	}
	return Token{Type: TokenEOL, Line: lexer.line, Column: lexer.column}
}

func (lexer *Lexer) lexUnknown() Token {

	startLine, startColumn := lexer.line, lexer.column
	word := ""
	for !unicode.IsSpace(lexer.peek()) && lexer.pos < len(lexer.input) {
		char := lexer.next()
		word = word + string(char)
	}
	return Token{
		Type:   TokenUnknown,
		Value:  word,
		Line:   startLine,
		Column: startColumn,
		Error:  "unknown token",
	}
}
