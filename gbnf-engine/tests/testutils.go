package tests

import (
	"gbnflsp/gbnf-engine/GBNFParser"
)

func CollectTokens(input string) []GBNFParser.Token {
	lexer := GBNFParser.NewLexer(input)
	return lexer.LexAllTokens()
}
