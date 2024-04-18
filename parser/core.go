package parser

import (
	"github.com/twolodzko/goer/parser/lexer"
	. "github.com/twolodzko/goer/types"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

// Peek at the next token without moving the cursor.
func (p *Parser) peek() (lexer.Token, bool) {
	if p.pos >= len(p.tokens) {
		return lexer.Token{}, false
	}
	return p.tokens[p.pos], true
}

// Return the next token, move the cursor.
func (p *Parser) pop() (lexer.Token, bool) {
	token, ok := p.peek()
	p.skip()
	return token, ok
}

// Skip the next token without returning it.
func (p *Parser) skip() {
	p.pos++
}

// Parse the string.
func Parse(input string) ([]Expr, error) {
	tokens, err := lexer.Tokenize(input)
	if err != nil {
		return nil, err
	}
	parser := Parser{tokens, 0}
	return parser.parseUntil(lexer.Dot)
}
