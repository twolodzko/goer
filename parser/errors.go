package parser

import (
	"fmt"

	"github.com/twolodzko/goer/parser/lexer"
)

type EoF struct{}

func (err EoF) Error() string {
	return "unexpected end of input"
}

type Unexpected struct {
	Token lexer.Token
}

func (err Unexpected) Error() string {
	return fmt.Sprintf("unexpected: %v", err.Token)
}

type Missing struct {
	Token lexer.TokenType
}

func (err Missing) Error() string {
	return fmt.Sprintf("missing: %v", err.Token)
}

type EmptyBody struct{}

func (err EmptyBody) Error() string {
	return "body of the expression cannot be empty"
}
