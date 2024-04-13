package lexer

import "fmt"

type EoF struct{}

func (err EoF) Error() string {
	return "unexpected end of input"
}

type Empty struct{}

func (err Empty) Error() string {
	return "empty token"
}

type Invalid struct{ Value string }

func (err Invalid) Error() string {
	return fmt.Sprintf("invalid token '%s'", err.Value)
}
