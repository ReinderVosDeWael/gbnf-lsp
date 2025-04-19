package lsp

import (
	"encoding/json"
	"fmt"
	"gbnflsp/gbnf-engine/GBNFParser"
	"os"
)

var shutdownRequested = false
var openFiles = map[string]*OpenFile{}

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
	Tokens       []*GBNFParser.Token
	AST          *GBNFParser.Node
	ParserErrors []*GBNFParser.ParseError
}

func (file OpenFile) GetRuleNames() []string {
	rules := recursiveGetRules(file.AST)
	uniqueRulesMap := make(map[string]bool)
	for _, rule := range rules {
		uniqueRulesMap[rule] = true
	}

	uniqueRules := []string{}
	for key := range uniqueRulesMap {
		uniqueRules = append(uniqueRules, key)
	}
	return uniqueRules

}

func recursiveGetRules(node *GBNFParser.Node) []string {
	rules := []string{}
	debugLogger.Print(node)
	if node != nil {
		if node.Token != nil {
			if node.Token.Type == GBNFParser.TokenIdentifier {
				debugLogger.Print("rule")
				debugLogger.Print(node.Token)
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
	debugLogger.Print("Tokenizing...")
	tokens, err := lexer.LexAllTokens()

	var lexerErrors string

	if err != nil {
		lexerErrors = "Could not tokenize the entire document, continuing with whatever was lexed"
		fmt.Fprintf(os.Stderr, lexerErrors)
	}
	debugLogger.Print("Parsing...")
	parser := GBNFParser.NewParser(tokens)
	ast, parseErrors := parser.ParseAllRules()
	debugLogger.Print("Finished parsing.")
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
