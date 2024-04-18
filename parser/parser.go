package parser

import (
	"strconv"

	"github.com/twolodzko/goer/parser/lexer"
	. "github.com/twolodzko/goer/types"
)

// Parse an expression. The expression can be an atom, integer, operation,
// if block, function definition, function call, etc. It is a standalone unit
// of code that can be evaluated.
func (p *Parser) parseExpr() (Expr, error) {
	expr, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	// maybe a call
	next, ok := p.peek()
	if ok && next.Type == lexer.BracketLeft {
		p.skip()
		switch expr.(type) {
		case Atom, Variable, Bracket:
			args, err := p.parseUntil(lexer.BracketRight)
			if err != nil {
				return nil, err
			}
			expr = Call{expr, args}
		default:
			// other things followed by a bracket does not make sense
			return nil, Unexpected{next}
		}
	}

	// maybe a binary operation
	next, ok = p.peek()
	if ok && isOperator(next) {
		p.skip()
		rhs, err := p.parseExpr()
		return newBinaryOperation(next.Value, expr, rhs), err
	}

	return expr, nil
}

// Parse a sequence of expressions, separated by "," and delimited by the `delim` token.
func (p *Parser) parseUntil(delim lexer.TokenType) ([]Expr, error) {
	var exprs []Expr

	// handle empty case
	token, ok := p.peek()
	if !ok {
		return exprs, Missing{delim}
	} else if token.Type == delim {
		p.skip()
		return exprs, nil
	}

	for {
		expr, err := p.parseExpr()
		if err != nil {
			return exprs, err
		}
		exprs = append(exprs, expr)

		// punctuation: "," to split and `delim` ends
		token, ok := p.pop()
		if !ok {
			return exprs, Missing{delim}
		}
		switch token.Type {
		case lexer.Comma:
			// skip
		case delim:
			return exprs, nil
		default:
			return exprs, Unexpected{token}
		}
	}
}

// Parse single term in an expression. The term can be a standalone unit of code,
// or it may have continuation (e.g. function call or a binary operation).
func (p *Parser) parseTerm() (Expr, error) {
	token, ok := p.pop()
	if !ok {
		return nil, EoF{}
	}

	switch token.Type {
	case lexer.Atom:
		return fromAtom(token), nil
	case lexer.Variable:
		return Variable(token.Value), nil
	case lexer.Dummy:
		return Dummy{}, nil
	case lexer.Number:
		num, err := strconv.Atoi(token.Value)
		return Int(num), err
	case lexer.String:
		val, err := strconv.Unquote(token.Value)
		return String(val), err
	case lexer.Operator:
		// it needs to be a unary operation
		if isOneOf(token.Value, "+", "-", "not") {
			rhs, err := p.parseTerm()
			return UnaryOperation{token.Value, rhs}, err
		} else {
			return nil, Unexpected{token}
		}
	case lexer.BraceLeft:
		exprs, err := p.parseUntil(lexer.BraceRight)
		return Tuple{exprs}, err
	case lexer.BracketLeft:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return Bracket{expr}, p.expect(lexer.BracketRight)
	case lexer.SuareBracketLeft:
		exprs, err := p.parseUntil(lexer.SquareBracketRight)
		return List{exprs}, err
	case lexer.Fun:
		var name string
		next, ok := p.peek()
		if !ok {
			return nil, EoF{}
		}
		if next.Type == lexer.Atom {
			name = next.Value
			p.skip()
		}
		branches, err := parseBranches(p, parseFunBranch)
		return Definition{name, branches}, err
	case lexer.If:
		branches, err := parseBranches(p, parseIfBranch)
		return If{branches}, err
	case lexer.Case:
		return p.parseCase()
	case lexer.Receive:
		return p.parseReceive()
	case lexer.Try:
		return p.parseTryRecover()
	default:
		return nil, Unexpected{token}
	}
}

// Token is a valid operator.
func isOperator(token lexer.Token) bool {
	if token.Type != lexer.Operator {
		return false
	}
	_, ok := operatorPrecedence[token.Value]
	return ok
}

// Expect the `token`, otherwise return an error.
func (p *Parser) expect(token lexer.TokenType) error {
	next, ok := p.pop()
	if !ok {
		return Missing{token}
	}
	switch next.Type {
	case token:
		return nil
	default:
		return Unexpected{next}
	}
}

// Initialize a `BinaryOperation` with correct operator precedence.
func newBinaryOperation(op string, lhs, rhs Expr) BinaryOperation {
	switch rhs := rhs.(type) {
	case BinaryOperation:
		if operatorPrecedence[op] <= operatorPrecedence[rhs.Op] {
			// correct precedence: 2 * (3 + 5) -> (2 * 3) + 5
			return BinaryOperation{rhs.Op, newBinaryOperation(op, lhs, rhs.Lhs), rhs.Rhs}
		}
	}
	return BinaryOperation{op, lhs, rhs}
}

// Check if the `value` is one of the following values.
func isOneOf[T comparable](value T, set ...T) bool {
	for _, x := range set {
		if value == x {
			return true
		}
	}
	return false
}

// Transform special atoms (booleans) to specific values.
func fromAtom(token lexer.Token) Expr {
	switch token.Value {
	case "true":
		return Bool(true)
	case "false":
		return Bool(false)
	default:
		return Atom(token.Value)
	}
}
