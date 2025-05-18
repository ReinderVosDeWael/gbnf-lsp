package lsp

import (
	"encoding/json"
	"fmt"
	"gbnflsp/gbnf-engine/GBNFParser"
)

type Diagnostic struct {
	Range    Range  `json:"range"`
	Message  string `json:"message"`
	Severity int    `json:"severity"` // 1 = Error, 2 = Warning
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
	file := openFiles[uri]
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

	rootRule := ruleMustIncludeRoot(uri)
	if rootRule != nil {
		diags = append(diags, rootRule)
	}

	return diags
}

func ruleMustIncludeRoot(uri string) *Diagnostic {
	file := openFiles[uri]
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
