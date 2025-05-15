package tests

import (
	"gbnflsp/gbnf-engine/GBNFParser"
	"strings"
	"testing"
)

func TestAssignmentToken(t *testing.T) {
	tokens := CollectTokens("rule ::= something")

	if tokens[1].Type != GBNFParser.TokenAssignment {
		t.Errorf("Expected TokenAssignment, not found")
	}
}

func TestStringToken(t *testing.T) {
	tokens := CollectTokens(`"hello world"`)

	if len(tokens) == 0 || tokens[0].Type != GBNFParser.TokenString || tokens[0].Value != "hello world" {
		t.Errorf("Expected TokenString with value 'hello world', got %+v", tokens)
	}
}

func TestUnterminatedString(t *testing.T) {
	token := CollectTokens(`"unterminated`)
	if !strings.Contains(token[0].Error, "unterminated string") {
		t.Errorf("Expected unterminated string error, got %v", token[0].Error)
	}
}

func TestRegexToken(t *testing.T) {
	tokens := CollectTokens(`[abc]`)

	if tokens[0].Type != GBNFParser.TokenRegexp || tokens[0].Value != "[abc]" {
		t.Errorf("Expected TokenRegexp with value '[abc]', got %+v", tokens[0])
	}
}

func TestUnterminatedRegex(t *testing.T) {
	token := CollectTokens(`[abc`)
	if !strings.Contains(token[0].Error, "unterminated regex") {
		t.Errorf("Expected unterminated regex error, got %v", token[0].Error)
	}
}

func TestCommentSkipping(t *testing.T) {
	tokens := CollectTokens("# comment\nname")

	if len(tokens) == 0 || tokens[0].Type != GBNFParser.TokenEOL || tokens[1].Type != GBNFParser.TokenIdentifier || tokens[1].Value != "name" {
		t.Errorf("Expected TokenIdentifier 'name' after comment, got %+v", tokens)
	}
}

func TestRangeToken(t *testing.T) {
	tokens := CollectTokens(`"abc"{1,2}`)

	if tokens[1].Type != GBNFParser.TokenRepeat || tokens[1].Value != "{1,2}" {
		t.Errorf("Expected TokenOperator '{1,2}', got %+v", tokens[1])
	}
}

func TestRangeTokenMissingSecond(t *testing.T) {
	tokens := CollectTokens(`"abc"{1,}`)

	if tokens[1].Type != GBNFParser.TokenRepeat || tokens[1].Value != "{1,}" {
		t.Errorf("Expected TokenOperator '{1,2}', got %+v", tokens[1])
	}
}

func TestInvalidRangeNonNumeric(t *testing.T) {
	token := CollectTokens(`"abc"{a,b}`)
	if !strings.Contains(token[1].Error, "must be numeric") {
		t.Errorf("Expected numeric range error, got %v", token[0].Error)
	}
}

func TestSingleRangeValue(t *testing.T) {
	tokens := CollectTokens(`"abc"{1}`)
	if tokens[1].Value != "{1}" {
		t.Errorf("Expected empty numeric value error, got %v", tokens[1])
	}
}

func TestSimpleOperator(t *testing.T) {
	tokens := CollectTokens(`"abc"*`)

	if len(tokens) != 2 {
		t.Fatalf("Expected 2 tokens, got %v", len(tokens))
	}
	if tokens[1].Type != GBNFParser.TokenOperator || tokens[1].Value != "*" {
		t.Errorf("Expected TokenOperator '*', got %+v", tokens[0])
	}
}

func TestIdentifierToken(t *testing.T) {
	tokens := CollectTokens("abc123")

	if tokens[0].Type != GBNFParser.TokenIdentifier || tokens[0].Value != "abc123" {
		t.Errorf("Expected TokenIdentifier 'abc123', got %+v", tokens[0])
	}
}

func TestInvalidIdentifierCharacter(t *testing.T) {
	token := CollectTokens("var$")
	if !strings.Contains(token[0].Error, "unknown variable name") {
		t.Errorf("Expected Unknown variable name error, got %v", token[0].Error)
	}
}

func TestNewlineToken(t *testing.T) {
	tokens := CollectTokens("\n")

	if tokens[0].Type != GBNFParser.TokenEOL {
		t.Errorf("Expected TokenEOL, got %+v", tokens[0])
	}
}

func TestMultipleTokensSequence(t *testing.T) {
	input := `
# comment
rule ::= "value" [a-z]{1,3} identifier*`
	tokens := CollectTokens(input)

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
	tokens := CollectTokens(input)

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
	tokens := CollectTokens(input)

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

func TestUnterminatedStringMultiline(t *testing.T) {
	input := `rule ::= "a
	word ::= "abc"`
	tokens := CollectTokens(input)

	expectedTypes := []GBNFParser.TokenType{
		GBNFParser.TokenIdentifier,
		GBNFParser.TokenAssignment,
		GBNFParser.TokenString,
		GBNFParser.TokenIdentifier,
		GBNFParser.TokenAssignment,
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
