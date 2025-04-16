package lsp

import (
	"encoding/json"
	"fmt"
	"os"
)

type DidOpenTextDocumentParams struct {
	TextDocument struct {
		URI        string `json:"uri"`
		LanguageID string `json:"languageId"`
		Version    int    `json:"version"`
		Text       string `json:"text"`
	} `json:"textDocument"`
}

func handleTextDocumentDidOpen(request Request) {
	params := request.Params
	var data DidOpenTextDocumentParams
	err := json.Unmarshal(params, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal didOpen: %v\nRaw: %s\n", err, params)
		return
	}

	newFile := TextToOpenFile(data.TextDocument.Text)
	openFiles[data.TextDocument.URI] = &newFile
	sendDiagnostics(data.TextDocument.URI)
}

type DidChangeTextDocumentParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

func handleTextDocumentDidChange(request Request) {
	// Currently only supports full sync
	params := request.Params
	var data DidChangeTextDocumentParams
	err := json.Unmarshal(params, &data)
	if err != nil {
		debugLogger.Print("Failed to unmarshal")
		fmt.Fprintf(os.Stderr, "Failed to unmarshal didChange: %v\nRaw: %s\n", err, params)
		return
	}

	debugLogger.Print("Text change: " + data.ContentChanges[0].Text)
	newFile := TextToOpenFile(data.ContentChanges[0].Text)
	openFiles[data.TextDocument.URI] = &newFile
	sendDiagnostics(data.TextDocument.URI)
}

func handleTextDocumentDidSave(request Request) {
	// No return
}

type DidCloseTextDocumentParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

func handleTextDocumentDidClose(request Request) {
	params := request.Params
	var data DidCloseTextDocumentParams
	err := json.Unmarshal(params, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal didClose: %v\nRaw: %s\n", err, params)
		return
	}
	delete(openFiles, data.TextDocument.URI)
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Message  string `json:"message"`
	Severity int    `json:"severity"` // 1 = Error, 2 = Warning
	Source   string `json:"source,omitempty"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

func createDiagnostics(uri string) []Diagnostic {
	file := openFiles[uri]
	errors := file.ParserErrors
	diags := []Diagnostic{}
	for _, err := range errors {
		diags = append(diags, Diagnostic{
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
	return diags
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
