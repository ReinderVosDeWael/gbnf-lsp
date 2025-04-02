package src

import (
	"fmt"
	"strconv"
	"strings"
)

type Node any

type RuleNode struct {
	Name       string
	Expression Node
}

type SequenceNode struct {
	Elements []Node
}

type AlternativeNode struct {
	Options []Node
}

type GroupNode struct {
	Expression Node
}

type TokenNode struct {
	Token Token
	min   int
	max   int
}

type Parser struct {
	tokens []Token
	pos    int
}

func (parser *Parser) peek() Token {
	if parser.pos < len(parser.tokens) {
		return parser.tokens[parser.pos]
	}
	return Token{Type: TokenEOL}
}

func (parser *Parser) next() Token {
	token := parser.peek()
	if token.Type != TokenEOL {
		parser.pos++
	}
	return token
}

func (parser *Parser) expect(expected TokenType) (Token, error) {
	token := parser.next()
	if token.Type != expected {
		return Token{}, fmt.Errorf("expected %v, got %v at line %d, col %d", expected, token.Type, token.Line, token.Column)
	}
	return token, nil
}

func (parser *Parser) ParseRule() (*RuleNode, error) {
	nameToken, err := parser.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}

	_, err = parser.expect(TokenAssignment)
	if err != nil {
		return nil, err
	}

	expression, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}
	return &RuleNode{Name: nameToken.Value, Expression: expression}, nil
}

func (parser *Parser) parseExpression() (*[]TokenNode, error) {
	token := parser.peek()
	nodes := []TokenNode{}
	for token.Type != TokenEOL {
		token = parser.next()
		switch token.Type {
		case TokenAssignment:
			return nil, fmt.Errorf("unexpected assignment at %d:%d", token.Line, token.Column)
		case TokenString, TokenRegexp, TokenIdentifier:
			nodes = append(nodes, TokenNode{Token: token, min: 1, max: 1})
		case TokenOperator:
			if len(nodes) == 0 {
				return nil, fmt.Errorf("misplaced operator token at %d:%d", token.Line, token.Column)
			}
			operaterNode, err := parser.parseOperator(&nodes[len(nodes)-1])
			if err != nil {
				return nil, err
			}
			nodes[len(nodes)-1] = operaterNode

		case TokenRepeat:
			if len(nodes) == 0 {
				return nil, fmt.Errorf("unexpected repeat at %d:%d", token.Line, token.Column)
			}
			repeatNode, err := parser.parseRepeat(&nodes[len(nodes)-1])
			if err != nil {
				return nil, err
			}
			nodes[len(nodes)] = repeatNode

		}
	}
	return &nodes, nil
}

func (parser *Parser) parseOperator(previousNode *TokenNode) (TokenNode, error) {
	token := parser.next()

	if token.Type != TokenOperator {
		return TokenNode{}, fmt.Errorf("invalid operator %s at %d:%d", token.Value, token.Line, token.Column)
	}
	if len(token.Value) == 0 {
		return TokenNode{}, fmt.Errorf("empty operator %s at %d:%d", token.Value, token.Line, token.Column)
	}

	var minRepeats int
	var maxRepeats int
	switch token.Value[0] {
	case '?':
		minRepeats = 0
		maxRepeats = 1
	case '*':
		minRepeats = 0
		maxRepeats = -1
	case '+':
		minRepeats = 1
		maxRepeats = -1

	default:
		return TokenNode{}, fmt.Errorf("invalid operator %s at %d:%d", token.Value, token.Line, token.Column)
	}
	switch previousNode.Token.Type {
	case TokenString, TokenRegexp, TokenExpression, TokenIdentifier:
		return TokenNode{Token: token, min: minRepeats, max: maxRepeats}, nil
	default:
		return TokenNode{}, fmt.Errorf("cannot apply operator to this token type at %d:%d", token.Line, token.Column)

	}
}

func (parser *Parser) parseRepeat(previousNode *TokenNode) (TokenNode, error) {
	token := parser.next()
	if token.Type != TokenRepeat {
		return TokenNode{}, fmt.Errorf("invalid repeat %s at %d:%d", token.Value, token.Line, token.Column)
	}

	parts := strings.Split(token.Value, ",")
	var min int
	var max int
	var err error
	if len(parts) == 1 {
		min, err = strconv.Atoi(parts[0])
		if err != nil {
			return TokenNode{}, fmt.Errorf("could not parse min at %d:%d", token.Line, token.Column)
		}
		max = min
	} else if len(parts) == 2 {
		min, err = strconv.Atoi(parts[0])
		if err != nil {
			return TokenNode{}, fmt.Errorf("could not parse min at %d:%d", token.Line, token.Column)
		}
		max, err = strconv.Atoi(parts[1])
		if err != nil {
			return TokenNode{}, fmt.Errorf("could not parse max at %d:%d", token.Line, token.Column)
		}
	} else {
		return TokenNode{}, fmt.Errorf("expected 1-2 repeat parts, found %d at %d:%d", len(parts), token.Line, token.Column)
	}

	return TokenNode{Token: previousNode.Token, min: min, max: max}, nil
}
