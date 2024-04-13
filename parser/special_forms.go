package parser

import (
	"github.com/twolodzko/goer/parser/lexer"
	. "github.com/twolodzko/goer/types"
)

// Parse the "case" statement.
func (p *Parser) parseCase() (Expr, error) {
	arg, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	err = p.expect(lexer.Of)
	if err != nil {
		return nil, err
	}
	branches, err := parseBranches(p, parsePatternBranch)
	return Case{arg, branches}, err
}

// Parse the "try ... recover ... end" block.
func (p *Parser) parseTryRecover() (Expr, error) {
	var (
		try TryRecover
		err error
	)

	try.Body, err = p.parseUntil(lexer.Recover)
	if err != nil {
		return nil, err
	}
	if len(try.Body) == 0 {
		return nil, Unexpected{lexer.Token{lexer.Recover, "recover"}}
	}

	try.Recover, err = p.parseUntil(lexer.End)
	if err != nil {
		return nil, err
	}
	if len(try.Recover) == 0 {
		return nil, Unexpected{lexer.Token{lexer.End, "end"}}
	}

	return try, nil
}

// Parse the "receive" statement.
func (p *Parser) parseReceive() (Expr, error) {
	var (
		receive Receive
		err     error
	)

	// after branch without any receive branches
	// see: https://www.erlang.org/doc/reference_manual/expressions#receive
	token, ok := p.peek()
	if !ok {
		return nil, EoF{}
	}
	if token.Type == lexer.After {
		p.skip()
		receive.After, err = parseIfBranch(p)
		if err != nil {
			return nil, err
		}
		return receive, p.expect(lexer.End)
	}

	for {
		branch, err := parsePatternBranch(p)
		if err != nil {
			return nil, err
		}
		receive.Branches = append(receive.Branches, branch)

		token, ok := p.pop()
		if !ok {
			return nil, EoF{}
		}
		switch token.Type {
		case lexer.Semicolon:
			// skip
		case lexer.End:
			return receive, nil
		case lexer.After:
			receive.After, err = parseAfterBranch(p)
			return receive, err
		default:
			return nil, Unexpected{token}
		}
	}
}

// Parse the "after" branch of the "receive" statement.
func parseAfterBranch(p *Parser) (IfBranch, error) {
	after, err := parseIfBranch(p)
	if err != nil {
		return IfBranch{}, err
	}
	return after, p.expect(lexer.End)
}

// Parse the function definition branch.
func parseFunBranch(p *Parser) (FunBranch, error) {
	var (
		branch FunBranch
		err    error
	)

	err = p.expect(lexer.BracketLeft)
	if err != nil {
		return branch, err
	}
	branch.Args, err = p.parseUntil(lexer.BracketRight)

	branch.Guards, err = p.maybeGuards()
	if err != nil {
		return branch, err
	}

	branch.Body, err = p.parseBranchBody()
	return branch, err
}

// Parse the simple `expr "->" body` branch.
func parseIfBranch(p *Parser) (IfBranch, error) {
	var (
		branch IfBranch
		err    error
	)

	branch.Cond, err = p.ParseExpr()
	if err != nil {
		return branch, err
	}

	err = p.expect(lexer.Arrow)
	if err != nil {
		return branch, err
	}

	branch.Body, err = p.parseBranchBody()
	return branch, err
}

// Parse the branch of the `expr [ "which" guards ]? "->" body` form.
func parsePatternBranch(p *Parser) (PatternBranch, error) {
	var (
		branch PatternBranch
		err    error
	)

	branch.Pattern, err = p.ParseExpr()
	if err != nil {
		return branch, err
	}

	branch.Guards, err = p.maybeGuards()
	if err != nil {
		return branch, err
	}

	branch.Body, err = p.parseBranchBody()
	return branch, err
}

// Parse individual branches (in "if", "case", or "fun" blocks) until the "end" token.
func parseBranches[T any](p *Parser, parse func(*Parser) (T, error)) ([]T, error) {
	var branches []T
	for {
		branch, err := parse(p)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branch)

		// check if this is the final branch
		token, ok := p.pop()
		if !ok {
			return nil, Missing{lexer.End}
		}
		switch token.Type {
		case lexer.Semicolon:
			// skip
		case lexer.End:
			return branches, nil
		default:
			return nil, Unexpected{token}
		}
	}
}

// Parse body of the branch until ";" or "end".
func (p *Parser) parseBranchBody() ([]Expr, error) {
	var body []Expr
	for {
		expr, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		body = append(body, expr)

		// punctuation: "," to split and "end" or ";" ends
		token, ok := p.peek()
		if !ok {
			return nil, Missing{lexer.End}
		}
		switch token.Type {
		case lexer.Comma:
			p.skip()
		case lexer.Semicolon, lexer.End, lexer.After:
			return body, nil
		default:
			return nil, Unexpected{token}
		}
	}
}

// Expect "->" or parse and return `"when" guards "->"`.
func (p *Parser) maybeGuards() ([]Expr, error) {
	token, ok := p.pop()
	if !ok {
		return nil, Missing{lexer.Arrow}
	}
	switch token.Type {
	case lexer.Arrow:
		return nil, nil
	case lexer.When:
		return p.parseUntil(lexer.Arrow)
	default:
		return nil, Unexpected{token}
	}
}
