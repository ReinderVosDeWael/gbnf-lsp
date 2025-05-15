package GBNFParser

import (
	"fmt"
	"strconv"
	"strings"
)

type NodeType int

const (
	NodeUnknown NodeType = iota
	NodeSubExpression
	NodeAlternative
	NodeToken
	NodeRoot
	NodeDeclaration
	NodeRepeat
)

func (t TokenType) String() string {
	switch t {
	case TokenAssignment:
		return "TokenAssignment"
	case TokenString:
		return "TokenString"
	case TokenRegexp:
		return "TokenRegexp"
	case TokenOperator:
		return "TokenOperator"
	case TokenAlternative:
		return "TokenAlternative"
	case TokenSubExpression:
		return "TokenSubExpression"
	case TokenIdentifier:
		return "TokenIdentifier"
	case TokenRepeat:
		return "TokenRepeat"
	case TokenEOL:
		return "TokenEOL"
	default:
		return "TokenUnknown"
	}
}

type ParseError struct {
	Message string
	Line    int
	Column  int
	Length  int
}

func NewParseError(msg string, token *Token, args ...any) *ParseError {
	return &ParseError{
		Message: fmt.Sprintf(msg, args...),
		Line:    token.Line,
		Column:  token.Column,
		Length:  len(token.Value),
	}
}

type Node struct {
	Min      int
	Max      int
	Children []*Node
	Token    *Token
	Type     NodeType
}

type Parser struct {
	Tokens []Token
	pos    int
}

func NewParser(tokens []Token) Parser {
	return Parser{
		Tokens: tokens,
		pos:    0,
	}
}

func (parser *Parser) ParseAllRules() (*Node, []*ParseError) {
	rules := []*Node{}
	errors := []*ParseError{}
	previousPos := 0
	for parser.pos < len(parser.Tokens) {
		newRule, err := parser.ParseRule()
		if previousPos == parser.pos {
			errors = append(errors, NewParseError("Parser entered a loop", parser.peek()))
			break
		}
		previousPos = parser.pos
		if err != nil {
			errors = append(errors, err)
		}
		if newRule != nil {
			rules = append(rules, newRule)
		}
	}
	root := Node{Type: NodeRoot, Children: rules}
	return &root, errors
}

func (parser *Parser) peek() *Token {
	if parser.pos < len(parser.Tokens) {
		return &parser.Tokens[parser.pos]
	}
	return &Token{Type: TokenEOL}
}

func (parser *Parser) next() *Token {
	token := parser.peek()
	if parser.pos < len(parser.Tokens) {
		parser.pos++
	}
	return token
}

func (parser *Parser) expect(expected TokenType) (*Token, *ParseError) {
	token := parser.next()

	if token.Error != "" {
		token := parser.next()
		return nil, &ParseError{
			Message: token.Error,
			Line:    token.Line,
			Column:  token.Column,
			Length:  len(token.Value),
		}
	}

	if token.Type != expected {
		return nil, &ParseError{
			Message: fmt.Sprintf("expected %v, got %v", expected, token.Type),
			Line:    token.Line,
			Column:  token.Column,
			Length:  len(token.Value),
		}
	}
	return token, nil
}

func (parser *Parser) ParseRule() (*Node, *ParseError) {

	if parser.peek().Type == TokenEOL {
		parser.next()
		return nil, nil
	}

	nameToken, err := parser.expect(TokenIdentifier)
	if err != nil {
		parser.forwardTillNextLine()
		return nil, err
	}

	_, err = parser.expect(TokenAssignment)
	if err != nil {
		parser.forwardTillNextLine()
		return nil, err
	}

	root := Node{Token: nameToken, Type: NodeDeclaration}
	children, err := parser.parseExpression()
	root.Children = children
	if err != nil {
		parser.forwardTillNextLine()
		return nil, err
	}
	return &root, nil
}

func (parser *Parser) forwardTillNextLine() {
	if parser.pos >= len(parser.Tokens) {
		return
	}
	current := parser.Tokens[parser.pos]
	for current.Line == parser.peek().Line && parser.pos < len(parser.Tokens) {
		parser.next()
	}

}

func (parser *Parser) parseExpression() ([]*Node, *ParseError) {
	nodes := []*Node{}
	lastTokenAlternative := false

Loop:
	for {
		token := parser.peek()
		if parser.peek().Error != "" {
			return nil, &ParseError{
				Message: token.Error,
				Line:    token.Line,
				Column:  token.Column,
				Length:  len(token.Value),
			}
		}

		if token.Type == TokenEOL && !lastTokenAlternative {
			if len(nodes) == 0 {
				return nil, NewParseError("empty expression", token)
			}
			break
		}
		if token.Type == TokenEOL {
			parser.next()
			continue
		}

		lastTokenAlternative = false
		token = parser.next()
		switch token.Type {

		case TokenUnknown:
			err := ParseError{
				Message: "unknown token",
				Line:    token.Line,
				Column:  token.Column,
				Length:  len(token.Value),
			}
			return nil, &err
		case TokenAssignment:
			err := ParseError{
				Message: "unexpected assignment",
				Line:    token.Line,
				Column:  token.Column,
				Length:  len(token.Value),
			}
			return nil, &err

		case TokenAlternative:
			// Alternatives are done after parsing the entire expression.
			lastTokenAlternative = true
			nodes = append(nodes, &Node{Type: NodeAlternative, Min: 1, Max: 1, Token: token})
		case TokenString, TokenRegexp, TokenIdentifier:
			nodes = append(nodes, &Node{Token: token, Min: 1, Max: 1, Type: NodeToken})

		case TokenOperator:
			if len(nodes) == 0 {
				err := ParseError{
					Message: "misplaced operator token",
					Line:    token.Line,
					Column:  token.Column,
					Length:  len(token.Value),
				}
				return nil, &err
			}
			operaterNode, err := parser.parseOperator(nodes[len(nodes)-1], token)
			if err != nil {
				return nil, err
			}
			nodes[len(nodes)-1] = operaterNode

		case TokenRepeat:
			if len(nodes) == 0 {
				return nil, NewParseError("unexpected repeat", token)
			}
			repeatNode, err := parser.parseRepeat(nodes[len(nodes)-1], token)
			if err != nil {
				return nil, err
			}
			nodes[len(nodes)-1] = repeatNode

		case TokenSubExpression:
			if token.Value == "(" {
				children, err := parser.parseExpression()
				if err != nil {
					return nil, err
				}
				newNode := Node{Type: NodeSubExpression}
				newNode.Children = children
				nodes = append(nodes, &newNode)
			} else {
				break Loop
			}
		}
	}

	nodes, err := parser.parseAlternatives(nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (parser *Parser) parseAlternatives(nodes []*Node) ([]*Node, *ParseError) {
	newNodes := []*Node{}
	for index := 0; index < len(nodes); index++ {
		node := nodes[index]
		if node.Type != NodeAlternative {
			newNodes = append(newNodes, node)
			continue
		}
		// Guard: at start or end
		if index == 0 || index == len(nodes)-1 {
			return []*Node{}, NewParseError("alternative found at start or end of expression", node.Token)
		}

		previousNode := newNodes[len(newNodes)-1]
		nextNode := nodes[index+1]

		// Guard: two alternatives in a row
		if nextNode.Type == NodeAlternative {
			return []*Node{}, NewParseError("cannot have two alternatives in succession", node.Token)
		}
		if previousNode.Type == NodeAlternative {
			previousNode.Children = append(previousNode.Children, nextNode)
		} else {
			node.Children = []*Node{previousNode, nextNode}
			newNodes[len(newNodes)-1] = node
		}
		index++
	}

	// Ensure end-of-line alternative is handled correctly.
	if parser.peek().Type == TokenEOL {
		parser.next()
	}
	return newNodes, nil
}

func (parser *Parser) parseOperator(previousNode *Node, token *Token) (*Node, *ParseError) {
	if token.Type != TokenOperator {
		return nil, NewParseError("invalid operator %s", token, token.Value)
	}
	if len(token.Value) == 0 {
		return nil, NewParseError("empty operator", token)
	}

	var minRepeats, maxRepeats int
	switch token.Value[0] {
	case '?':
		minRepeats, maxRepeats = 0, 1
	case '*':
		minRepeats, maxRepeats = 0, -1
	case '+':
		minRepeats, maxRepeats = 1, -1
	default:
		return nil, NewParseError("invalid operator %s", token, token.Value)
	}

	switch previousNode.Type {
	case NodeSubExpression, NodeToken:
		return &Node{
			Token:    token,
			Min:      minRepeats,
			Max:      maxRepeats,
			Type:     NodeRepeat,
			Children: []*Node{previousNode},
		}, nil
	default:
		return nil, NewParseError("cannot apply operator to this token type", token)
	}
}

func (parser *Parser) parseRepeat(previousNode *Node, token *Token) (*Node, *ParseError) {
	if token.Type != TokenRepeat {
		return nil, NewParseError("invalid repeat %s", token, token.Value)
	}

	parts := strings.Split(token.Value[1:len(token.Value)-1], ",")

	var min, max int
	var err error

	if len(parts) == 1 {
		min, err = strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, NewParseError("could not parse min %q", token, parts[0])
		}
		max = min

	} else if len(parts) == 2 {
		minStr := strings.TrimSpace(parts[0])
		maxStr := strings.TrimSpace(parts[1])

		min, err = strconv.Atoi(minStr)
		if err != nil {
			return nil, NewParseError("could not parse min %q", token, minStr)
		}

		if len(maxStr) == 0 {
			max = -1
		} else {
			max, err = strconv.Atoi(maxStr)

			if err != nil {
				return nil, NewParseError("could not parse max %q", token, maxStr)
			}
		}
	} else {
		return nil, NewParseError("expected 1 or 2 repeat parts, got %d", token, len(parts))
	}

	return &Node{
		Token:    token,
		Min:      min,
		Max:      max,
		Type:     NodeRepeat,
		Children: []*Node{previousNode},
	}, nil
}
