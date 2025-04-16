package tests

import (
	"gbnflsp/go-src/GBNFParser"
	"strings"
	"testing"
)

func TestAssignmentToken(t *testing.T) {
	tokens, err := CollectTokens("rule ::= something")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Type != GBNFParser.TokenAssignment {
		t.Errorf("Expected TokenAssignment, not found")
	}
}

func TestStringToken(t *testing.T) {
	tokens, err := CollectTokens(`"hello world"`)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) == 0 || tokens[0].Type != GBNFParser.TokenString || tokens[0].Value != "hello world" {
		t.Errorf("Expected TokenString with value 'hello world', got %+v", tokens)
	}
}

func TestUnterminatedString(t *testing.T) {
	_, err := CollectTokens(`"unterminated`)
	if err == nil || !strings.Contains(err[0].Message, "unterminated string") {
		t.Errorf("Expected unterminated string error, got %v", err)
	}
}

func TestRegexToken(t *testing.T) {
	tokens, err := CollectTokens(`[abc]`)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != GBNFParser.TokenRegexp || tokens[0].Value != "[abc]" {
		t.Errorf("Expected TokenRegexp with value '[abc]', got %+v", tokens[0])
	}
}

func TestUnterminatedRegex(t *testing.T) {
	_, err := CollectTokens(`[abc`)
	if err == nil || !strings.Contains(err[0].Message, "unterminated regex") {
		t.Errorf("Expected unterminated regex error, got %v", err)
	}
}

func TestCommentSkipping(t *testing.T) {
	tokens, err := CollectTokens("# comment\nname")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) == 0 || tokens[0].Type != GBNFParser.TokenEOL || tokens[1].Type != GBNFParser.TokenIdentifier || tokens[1].Value != "name" {
		t.Errorf("Expected TokenIdentifier 'name' after comment, got %+v", tokens)
	}
}

func TestRangeToken(t *testing.T) {
	tokens, err := CollectTokens(`"abc"{1,2}`)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Type != GBNFParser.TokenRepeat || tokens[1].Value != "{1,2}" {
		t.Errorf("Expected TokenOperator '{1,2}', got %+v", tokens[1])
	}
}

func TestInvalidRangeNonNumeric(t *testing.T) {
	_, err := CollectTokens(`"abc"{a,b}`)
	if err == nil || !strings.Contains(err[0].Message, "must be numeric") {
		t.Errorf("Expected numeric range error, got %v", err)
	}
}

func TestEmptyRangeValue(t *testing.T) {
	_, err := CollectTokens(`"abc"{1,}`)
	if err == nil || !strings.Contains(err[0].Message, "empty numeric value") {
		t.Errorf("Expected empty numeric value error, got %v", err)
	}
}

func TestSingleRangeValue(t *testing.T) {
	tokens, err := CollectTokens(`"abc"{1}`)
	if err != nil || tokens[1].Value != "{1}" {
		t.Errorf("Expected empty numeric value error, got %v", err)
	}
}

func TestSimpleOperator(t *testing.T) {
	tokens, err := CollectTokens(`"abc"*`)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("Expected 2 tokens, got %v", len(tokens))
	}
	if tokens[1].Type != GBNFParser.TokenOperator || tokens[1].Value != "*" {
		t.Errorf("Expected TokenOperator '*', got %+v", tokens[0])
	}
}

func TestIdentifierToken(t *testing.T) {
	tokens, err := CollectTokens("abc123")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != GBNFParser.TokenIdentifier || tokens[0].Value != "abc123" {
		t.Errorf("Expected TokenIdentifier 'abc123', got %+v", tokens[0])
	}
}

func TestInvalidIdentifierCharacter(t *testing.T) {
	_, err := CollectTokens("var$")
	if err == nil || !strings.Contains(err[0].Message, "unknown variable name") {
		t.Errorf("Expected Unknown variable name error, got %v", err)
	}
}

func TestNewlineToken(t *testing.T) {
	tokens, err := CollectTokens("\n")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Type != GBNFParser.TokenEOL {
		t.Errorf("Expected TokenEOL, got %+v", tokens[0])
	}
}

func TestMultipleTokensSequence(t *testing.T) {
	input := `
# comment
rule ::= "value" [a-z]{1,3} identifier*`
	tokens, err := CollectTokens(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []GBNFParser.TokenType{
		GBNFParser.TokenEOL,
		GBNFParser.TokenEOL,
		GBNFParser.TokenIdentifier,
		GBNFParser.TokenAssignment,
		GBNFParser.TokenString,
		GBNFParser.TokenRegexp,
		GBNFParser.TokenRepeat,
		GBNFParser.TokenIdentifier,
		GBNFParser.TokenOperator,
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

func TestLexerNestedExpressions(t *testing.T) {
	input := `rule ::= ("a" | "b")*`
	tokens, err := CollectTokens(input)

	if err != nil {
		t.Fatalf("Unexpected lexer error: %v", err)
	}

	expectedTypes := []GBNFParser.TokenType{
		GBNFParser.TokenIdentifier,    // rule
		GBNFParser.TokenAssignment,    // ::=
		GBNFParser.TokenSubExpression, // (
		GBNFParser.TokenString,        // "a"
		GBNFParser.TokenAlternative,   // |
		GBNFParser.TokenString,        // "b"
		GBNFParser.TokenSubExpression, // )
		GBNFParser.TokenOperator,      // *
	}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("Expected %d tokens, got %d: %+v", len(expectedTypes), len(tokens), tokens)
	}

	for i, tok := range tokens {
		if tok.Type != expectedTypes[i] {
			t.Errorf("Token %d: expected type %v, got %v (value: %q)", i, expectedTypes[i], tok.Type, tok.Value)
		}
	}
}

func TestTokenUnknown(t *testing.T) {
	input := `rule := "a"`
	tokens, err := CollectTokens(input)

	if err == nil || err[0].Message != "unknown token" {
		t.Errorf("Did not receive the expected error")
	}

	expectedTypes := []GBNFParser.TokenType{
		GBNFParser.TokenIdentifier, // rule
		GBNFParser.TokenUnknown,    // :=
		GBNFParser.TokenString,
	}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("Expected %d tokens, got %d: %+v", len(expectedTypes), len(tokens), tokens)
	}

	for i, tok := range tokens {
		if tok.Type != expectedTypes[i] {
			t.Errorf("Token %d: expected type %v, got %v (value: %q)", i, expectedTypes[i], tok.Type, tok.Value)
		}
	}
}
