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

func TestRuleMustDefineRoot(t *testing.T) {
	text := `smoot ::= arg`

	openFile := lsp.TextToOpenFile(text)
	uri := "fake"
	lsp.OpenFiles[uri] = &openFile

	diagnostics := lsp.RuleMustIncludeRoot(uri)

	if diagnostics == nil {
		t.Fatalf("Expected 1 diagnostic, got nil")
	}

	if diagnostics.Message != "No `root` node found." {
		t.Errorf("Unexpected diagnostic message: %s", diagnostics.Message)
	}
	if diagnostics.Range.Start.Line != 0 || diagnostics.Range.Start.Character != 0 {
		t.Errorf("Unexpected diagnostic position: %+v", diagnostics.Range.Start)
	}
}

func TestRuleMustDefineAllVariablesUndefinedReference(t *testing.T) {
	text := `root ::= arg`

	openFile := lsp.TextToOpenFile(text)
	uri := "fake"
	lsp.OpenFiles[uri] = &openFile

	diagnostics := lsp.RuleMustDefineAllVariables(uri)

	if len(diagnostics) != 1 {
		t.Fatalf("Expected 1 diagnostic, got %d", len(diagnostics))
	}

	diag := diagnostics[0]
	if diag.Message != "Variable `arg` undefined." {
		t.Errorf("Unexpected diagnostic message: %s", diag.Message)
	}
	if diag.Range.Start.Line != 0 || diag.Range.Start.Character != 9 {
		t.Errorf("Unexpected diagnostic position: %+v", diag.Range.Start)
	}
}
