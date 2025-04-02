package src

import (
	"io"
	"strings"
	"testing"
)

func collectTokens(input string) ([]Token, error) {
	lexer := NewLexer(input)
	var tokens []Token
	for {
		tok, err := lexer.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return tokens, err
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

func TestAssignmentToken(t *testing.T) {
	tokens, err := collectTokens("rule ::= something")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Type != TokenAssignment {
		t.Errorf("Expected TokenAssignment, not found")
	}
}

func TestStringToken(t *testing.T) {
	tokens, err := collectTokens(`"hello world"`)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) == 0 || tokens[0].Type != TokenString || tokens[0].Value != "hello world" {
		t.Errorf("Expected TokenString with value 'hello world', got %+v", tokens)
	}
}

func TestUnterminatedString(t *testing.T) {
	_, err := collectTokens(`"unterminated`)
	if err == nil || !strings.Contains(err.Error(), "unterminated string") {
		t.Errorf("Expected unterminated string error, got %v", err)
	}
}

func TestRegexToken(t *testing.T) {
	tokens, err := collectTokens(`[abc]`)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != TokenRegexp || tokens[0].Value != "[abc]" {
		t.Errorf("Expected TokenRegexp with value '[abc]', got %+v", tokens[0])
	}
}

func TestUnterminatedRegex(t *testing.T) {
	_, err := collectTokens(`[abc`)
	if err == nil || !strings.Contains(err.Error(), "unterminated regex") {
		t.Errorf("Expected unterminated regex error, got %v", err)
	}
}

func TestCommentSkipping(t *testing.T) {
	tokens, err := collectTokens("# comment\nname")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) == 0 || tokens[0].Type != TokenIdentifier || tokens[0].Value != "name" {
		t.Errorf("Expected TokenIdentifier 'name' after comment, got %+v", tokens)
	}
}

func TestRangeToken(t *testing.T) {
	tokens, err := collectTokens(`"abc"{1,2}`)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Type != TokenOperator || tokens[1].Value != "{1,2}" {
		t.Errorf("Expected TokenOperator '{1,2}', got %+v", tokens[1])
	}
}

func TestInvalidRangeNonNumeric(t *testing.T) {
	_, err := collectTokens(`"abc"{a,b}`)
	if err == nil || !strings.Contains(err.Error(), "must be numeric") {
		t.Errorf("Expected numeric range error, got %v", err)
	}
}

func TestEmptyRangeValue(t *testing.T) {
	_, err := collectTokens(`"abc"{1,}`)
	if err == nil || !strings.Contains(err.Error(), "empty numeric value") {
		t.Errorf("Expected empty numeric value error, got %v", err)
	}
}

func TestSingleRangeValue(t *testing.T) {
	tokens, err := collectTokens(`"abc"{1}`)
	if err != nil || tokens[1].Value != "{1}" {
		t.Errorf("Expected empty numeric value error, got %v", err)
	}
}

func TestSimpleOperator(t *testing.T) {
	tokens, err := collectTokens(`"abc"*`)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Type != TokenOperator || tokens[1].Value != "*" {
		t.Errorf("Expected TokenOperator '*', got %+v", tokens[0])
	}
}

func TestIdentifierToken(t *testing.T) {
	tokens, err := collectTokens("abc123")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != TokenIdentifier || tokens[0].Value != "abc123" {
		t.Errorf("Expected TokenIdentifier 'abc123', got %+v", tokens[0])
	}
}

func TestInvalidIdentifierCharacter(t *testing.T) {
	_, err := collectTokens("var$")
	if err == nil || !strings.Contains(err.Error(), "unknown variable name") {
		t.Errorf("Expected Unknown variable name error, got %v", err)
	}
}

func TestNewlineToken(t *testing.T) {
	tokens, err := collectTokens("\n")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != TokenEOL {
		t.Errorf("Expected TokenEOL, got %+v", tokens[0])
	}
}

func TestMultipleTokensSequence(t *testing.T) {
	input := `
# comment
rule ::= "value" [a-z]{1,3} identifier*`
	tokens, err := collectTokens(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []TokenType{
		TokenEOL,
		TokenIdentifier,
		TokenAssignment,
		TokenString,
		TokenRegexp,
		TokenOperator,
		TokenIdentifier,
		TokenOperator,
	}

	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Expected token type %v at position %d, got %v", expected[i], i, tok.Type)
		}
	}
}
