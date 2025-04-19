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
		debugLogger.Printf("Found error: %v", err)
		diags = append(diags, Diagnostic{
			Range: Range{
				Start: Position{
					Line:      err.Line,
					Character: err.Column - 1, // VSCode starts the line _after_ this character.
				},
				End: Position{
					Line:      err.Line,
					Character: err.Column + err.Length - 1,
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

type CompletionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	} `json:"position"`
}

type CompletionItem struct {
	Label      string `json:"label"`
	Kind       int    `json:"kind,omitempty"` // 6 = Variable
	Detail     string `json:"detail,omitempty"`
	InsertText string `json:"insertText,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

func handleTextDocumentCompletion(request Request) {
	var params CompletionParams
	err := json.Unmarshal(request.Params, &params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal completion params: %v\n", err)
		return
	}

	file := openFiles[params.TextDocument.URI]
	var items []CompletionItem

	for _, name := range file.GetRuleNames() {
		items = append(items, CompletionItem{
			Label:      name,
			Kind:       6,
			Detail:     "Rule",
			InsertText: name,
		})
	}

	result := CompletionList{
		IsIncomplete: false,
		Items:        items,
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      request.ID,
		"result":  result,
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), data)
}
