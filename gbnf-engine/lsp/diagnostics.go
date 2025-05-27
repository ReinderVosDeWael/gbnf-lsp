package lsp

import (
	"encoding/json"
	"fmt"
	"gbnflsp/gbnf-engine/GBNFParser"
	"slices"
)

type Diagnostic struct {
	Range    Range  `json:"range"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
	Source   string `json:"source,omitempty"`
}

type PublishDiagnosticsParams struct {
	URI         string        `json:"uri"`
	Diagnostics []*Diagnostic `json:"diagnostics"`
}

func sendDiagnostics(uri string) {
	diags := createDiagnostics(uri)
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "textDocument/publishDiagnostics",
		"params": PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diags,
		},
	}

	data, _ := json.Marshal(msg)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), data)
}

func createDiagnostics(uri string) []*Diagnostic {
	file := OpenFiles[uri]
	errors := file.ParserErrors
	diags := []*Diagnostic{}
	for _, err := range errors {
		debugLogger.Printf("Found error: %v", err)
		diags = append(diags, &Diagnostic{
			Range: Range{
				Start: Position{
					Line:      err.Line,
					Character: err.Column,
				},
				End: Position{
					Line:      err.Line,
					Character: err.Column + err.Length,
				},
			},
			Message:  err.Message,
			Severity: 1,
			Source:   "gnbf-lsp",
		})
	}

	rootRule := RuleMustIncludeRoot(uri)
	if rootRule != nil {
		diags = append(diags, rootRule)
	}
	diags = append(diags, RuleMustDefineAllVariables(uri)...)

	return diags
}

func RuleMustIncludeRoot(uri string) *Diagnostic {
	file := OpenFiles[uri]
	for _, node := range file.AST.Children {
		if node.Type == GBNFParser.NodeDeclaration && node.Token.Value == "root" {
			return nil
		}
	}
	return &Diagnostic{
		Range: Range{
			Start: Position{
				Line: 0, Character: 0,
			},
			End: Position{Line: 0, Character: 0},
		},
		Message:  "No `root` node found.",
		Severity: 1,
		Source:   "gbnf-lsp",
	}
}

func RuleMustDefineAllVariables(uri string) []*Diagnostic {
	file := OpenFiles[uri]
	nodeNames := []string{}
	for _, node := range file.AST.Children {
		if node.Type == GBNFParser.NodeDeclaration {
			nodeNames = append(nodeNames, node.Token.Value)
		}
	}
	undefinedNodes := []*Diagnostic{}
	for _, node := range file.AST.Children {
		undefinedNodes = append(undefinedNodes, recursiveUndefinedNodeSearch(node, nodeNames)...)
	}
	return undefinedNodes
}

func recursiveUndefinedNodeSearch(node *GBNFParser.Node, targetNames []string) []*Diagnostic {
	if node == nil || node.Token == nil {
		return nil
	}
	undefinedNodes := []*Diagnostic{}
	if node.Token.Type == GBNFParser.TokenIdentifier && !slices.Contains(targetNames, node.Token.Value) {
		undefinedNodes = append(undefinedNodes, &Diagnostic{
			Range: Range{
				Start: Position{
					Line: node.Token.Line, Character: node.Token.Column,
				},
				End: Position{Line: node.Token.Line, Character: node.Token.Column + len(node.Token.Value)},
			},
			Message:  "Variable `" + node.Token.Value + "` undefined.",
			Severity: 1,
			Source:   "gbnf-lsp",
		})
	}
	for _, child := range node.Children {
		if child != nil {
			undefinedNodes = append(undefinedNodes, recursiveUndefinedNodeSearch(child, targetNames)...)
		}
	}
	return undefinedNodes

}
