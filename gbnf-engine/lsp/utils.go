package lsp

import (
	"encoding/json"
	"fmt"
	"gbnflsp/gbnf-engine/GBNFParser"
	"os"
)

var shutdownRequested = false
var OpenFiles = map[string]*OpenFile{}

type Request struct {
	Jsonrpc string
	ID      interface{}
	Method  string
	Params  json.RawMessage
}

type Response struct {
	Jsonrpc string
	ID      interface{}
	Result  interface{}
	Error   interface{}
}

type OpenFile struct {
	Text         string
	Tokens       []GBNFParser.Token
	AST          *GBNFParser.Node
	ParserErrors []*GBNFParser.ParseError
}

func (file OpenFile) GetRuleNames() []string {
	rules := recursiveGetRules(file.AST)
	seen := make(map[string]bool)
	uniqueRules := []string{}

	for _, rule := range rules {
		if !seen[rule] {
			seen[rule] = true
			uniqueRules = append(uniqueRules, rule)
		}
	}

	return uniqueRules
}
func recursiveGetRules(node *GBNFParser.Node) []string {
	rules := []string{}
	if node != nil {
		if node.Token != nil {
			if node.Token.Type == GBNFParser.TokenIdentifier {
				rules = append(rules, node.Token.Value)
			}
		}
		for _, child := range node.Children {
			rules = append(rules, recursiveGetRules(child)...)
		}
	}
	return rules
}

func TextToOpenFile(text string) OpenFile {
	lexer := GBNFParser.NewLexer(text)
	tokens := lexer.LexAllTokens()
	parser := GBNFParser.NewParser(tokens)
	ast, parseErrors := parser.ParseAllRules()
	return OpenFile{
		Text:         text,
		Tokens:       tokens,
		AST:          ast,
		ParserErrors: parseErrors,
	}
}

func sendResponse(id interface{}, result interface{}) {
	debugLogger.Printf("Sending response %v", id)
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create response with %v and %v.", id, result)
	}
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), data)
}

func sendError(id interface{}, code int, message string) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	data, _ := json.Marshal(response)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), data)
}
