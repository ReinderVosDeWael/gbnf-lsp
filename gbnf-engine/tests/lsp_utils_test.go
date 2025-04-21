package tests

import (
	"testing"

	"gbnflsp/gbnf-engine/GBNFParser"
	"gbnflsp/gbnf-engine/lsp"
)

func makeToken(value string, ttype GBNFParser.TokenType) *GBNFParser.Token {
	return &GBNFParser.Token{
		Type:   ttype,
		Value:  value,
		Line:   1,
		Column: 1,
	}
}

func TestGetRuleNamesNilAST(t *testing.T) {
	file := lsp.OpenFile{AST: nil}
	rules := file.GetRuleNames()
	if len(rules) != 0 {
		t.Errorf("Expected no rules, got %v", rules)
	}
}

func TestGetRuleNamesSingleAssignment(t *testing.T) {
	node := &GBNFParser.Node{
		Token: makeToken("rule1", GBNFParser.TokenIdentifier),
	}
	file := lsp.OpenFile{AST: node}
	rules := file.GetRuleNames()
	if len(rules) != 1 || rules[0] != "rule1" {
		t.Errorf("Expected [rule1], got %v", rules)
	}
}

func TestGetRuleNamesNestedAssignments(t *testing.T) {
	child1 := &GBNFParser.Node{
		Token: makeToken("rule2", GBNFParser.TokenIdentifier),
	}
	child2 := &GBNFParser.Node{
		Token: makeToken("rule3", GBNFParser.TokenIdentifier),
	}
	root := &GBNFParser.Node{
		Token:    makeToken("rule1", GBNFParser.TokenIdentifier),
		Children: []*GBNFParser.Node{child1, child2},
	}
	file := lsp.OpenFile{AST: root}
	rules := file.GetRuleNames()
	expected := []string{"rule1", "rule2", "rule3"}

	if len(rules) != len(expected) {
		t.Fatalf("Expected %d rules, got %v", len(expected), rules)
	}
	for i := range expected {
		if rules[i] != expected[i] {
			t.Errorf("Expected rule %q at index %d, got %q", expected[i], i, rules[i])
		}
	}
}
