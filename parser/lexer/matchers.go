package lexer

import "unicode"

// Character that can be a part of atom or variable name.
func isName(r rune) bool {
	// see: https://www.erlang.org/doc/reference_manual/expressions
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' || r == '@'
}

// The rune can be part of the operator name.
func isOperator(r rune) bool {
	switch r {
	case '+', '-', '*', '/', '=', '<', '>', ':', '!':
		return true
	default:
		return false
	}
}

// Match a single-character token.
func literalToken(r rune) (Token, error) {
	var typ TokenType
	switch r {
	case '.':
		typ = Dot
	case ',':
		typ = Comma
	case ';':
		typ = Semicolon
	case '(':
		typ = BracketLeft
	case ')':
		typ = BracketRight
	case '{':
		typ = BraceLeft
	case '}':
		typ = BraceRight
	case '[':
		typ = SuareBracketLeft
	case ']':
		typ = SquareBracketRight
	case '_':
		typ = Dummy
	default:
		return Token{}, Invalid{string(r)}
	}
	return Token{typ, string(r)}, nil
}

func atomType(s string) TokenType {
	switch s {
	case "not", "rem", "div", "and", "or":
		return Operator
	case "fun":
		return Fun
	case "if":
		return If
	case "case":
		return Case
	case "of":
		return Of
	case "receive":
		return Receive
	case "after":
		return After
	case "end":
		return End
	case "when":
		return When
	case "try":
		return Try
	case "recover":
		return Recover
	default:
		return Atom
	}
}

func operatorType(s string) TokenType {
	switch s {
	case "->":
		return Arrow
	default:
		return Operator
	}
}
