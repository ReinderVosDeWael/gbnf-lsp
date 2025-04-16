package tests

import (
	"gbnflsp/go-src/GBNFParser"
)

func CollectTokens(input string) ([]*GBNFParser.Token, []*GBNFParser.LexerError) {
	lexer := GBNFParser.NewLexer(input)
	return lexer.LexAllTokens()
}
