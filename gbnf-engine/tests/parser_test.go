package tests

import (
	"gbnflsp/gbnf-engine/GBNFParser"
	"testing"
)

func TestParserSimpleRule(t *testing.T) {
	tokens := CollectTokens(`name ::= "value"`)
	parser := GBNFParser.Parser{Tokens: tokens}
	node, err := parser.ParseRule()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if node.Token.Value != "name" || node.Type != GBNFParser.NodeDeclaration {
		t.Errorf("Expected NodeRoot with name 'name', got %+v", node)
	}

	if len(node.Children) != 1 || node.Children[0].Token.Value != "value" {
		t.Errorf("Expected child node with value 'value', got %+v", node.Children)
	}
}

func TestParserOperatorRule(t *testing.T) {
	tokens := CollectTokens(`digits ::= [0-9]+`)
	parser := GBNFParser.Parser{Tokens: tokens}
	node, err := parser.ParseRule()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if node.Children[0].Min != 1 || node.Children[0].Max != -1 {
		t.Errorf("Expected repetition '+' operator, got %+v", node.Children[0])
	}
}

func TestParserAlternativeRule(t *testing.T) {
	tokens := CollectTokens(`rule ::= "yes" | "no"`)
	parser := GBNFParser.Parser{Tokens: tokens}
	node, err := parser.ParseRule()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(node.Children) != 1 || node.Children[0].Type != GBNFParser.NodeAlternative {
		t.Errorf("Expected NodeAlternative, got %+v", node.Children[0])
	}

	alts := node.Children[0].Children
	expected := []string{"yes", "no"}
	for i, alt := range alts {
		if alt.Token.Value != expected[i] {
			t.Errorf("Expected alternative '%s', got '%s'", expected[i], alt.Token.Value)
		}
	}
}

func TestParserAlternativeRuleMultiline(t *testing.T) {
	tokens := CollectTokens(`rule ::= "yes" | 
										 "no"`)
	parser := GBNFParser.Parser{Tokens: tokens}
	node, err := parser.ParseRule()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(node.Children) != 1 || node.Children[0].Type != GBNFParser.NodeAlternative {
		t.Errorf("Expected NodeAlternative, got %+v", node.Children[0])
	}

	alts := node.Children[0].Children
	expected := []string{"yes", "no"}
	for i, alt := range alts {
		if alt.Token.Value != expected[i] {
			t.Errorf("Expected alternative '%s', got '%s'", expected[i], alt.Token.Value)
		}
	}
}

func TestParserNestedExpressions(t *testing.T) {
	tokens := CollectTokens(`rule ::= ("a" | "b")*`)
	parser := GBNFParser.Parser{Tokens: tokens}
	node, err := parser.ParseRule()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(node.Children) != 1 || node.Children[0].Min != 0 || node.Children[0].Max != -1 {
		t.Errorf("Expected repeat node with '*', got %+v", node.Children[0])
	}
}

func TestParserInvalidAlternativePosition(t *testing.T) {
	tokens := CollectTokens(`rule ::= | "a"`)
	parser := GBNFParser.Parser{Tokens: tokens}
	_, err := parser.ParseRule()

	if err == nil || err.Message != "alternative found at start or end of expression" {
		t.Errorf("Expected alternative position error, got %v", err)
	}
}

func TestParserUnexpectedAssignment(t *testing.T) {
	tokens := CollectTokens(`rule ::= "a" ::= "b"`)
	parser := GBNFParser.Parser{Tokens: tokens}
	_, err := parser.ParseRule()

	if err == nil || err.Message != "unexpected assignment" {
		t.Errorf("Expected unexpected assignment error, got %v", err)
	}
}

func TestParserRepeatToken(t *testing.T) {
	tokens := CollectTokens(`letters ::= [a-z]{2,4}`)
	parser := GBNFParser.Parser{Tokens: tokens}
	node, err := parser.ParseRule()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	child := node.Children[0]
	if child.Min != 2 || child.Max != 4 {
		t.Errorf("Expected repeat {2,4}, got Min:%d Max:%d", child.Min, child.Max)
	}
}

func TestParserMultipleTokensSequence(t *testing.T) {
	input := `rule ::= "a" [0-9]? identifier+`
	tokens := CollectTokens(input)
	parser := GBNFParser.Parser{Tokens: tokens}
	node, err := parser.ParseRule()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedTypes := []GBNFParser.TokenType{GBNFParser.TokenString, GBNFParser.TokenOperator, GBNFParser.TokenOperator}
	if len(node.Children) != len(expectedTypes) {
		t.Fatalf("Expected %d children, got %d", len(expectedTypes), len(node.Children))
	}

	for i, child := range node.Children {
		if child.Token.Type != expectedTypes[i] {
			t.Errorf("Expected token type %v at position %d, got %v", expectedTypes[i], i, child.Token.Type)
		}
	}
}
