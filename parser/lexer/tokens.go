package lexer

type TokenType int

const (
	Atom               TokenType = iota // starts with a lowercase character
	Variable                            // starts with an uppercase character
	Dummy                               // starts with "_"
	Number                              // integer or float
	String                              // string
	Operator                            // operator, e.g. "+", "*", "==", "!", "not", ...
	BracketLeft                         // "("
	BracketRight                        // ")"
	BraceLeft                           // "{"
	BraceRight                          // "}"
	SuareBracketLeft                    // "["
	SquareBracketRight                  // "]"
	Comma                               // ","
	Semicolon                           // ";"
	Dot                                 // "."
	Arrow                               // "->"
	If                                  // "if"
	End                                 // "end"
	Case                                // "case"
	Of                                  // "of"
	Fun                                 // "fun"
	When                                // "when"
	Receive                             // "receive"
	After                               // "after"
	Try                                 // "try"
	Recover                             // "recover"
)

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) IsNil() bool {
	return t == Token{}
}

func (t Token) String() string {
	return t.Value
}

func (t TokenType) String() string {
	switch t {
	case Atom:
		return "atom"
	case Variable:
		return "var"
	case Dummy:
		return "_"
	case String:
		return "string"
	case Number:
		return "num"
	case Operator:
		return "op"
	case BracketLeft:
		return "("
	case BracketRight:
		return ")"
	case BraceLeft:
		return "{"
	case BraceRight:
		return "}"
	case SuareBracketLeft:
		return "["
	case SquareBracketRight:
		return "]"
	case Comma:
		return ","
	case Semicolon:
		return ";"
	case Dot:
		return "."
	case Arrow:
		return "->"
	case If:
		return "if"
	case End:
		return "end"
	case Case:
		return "case"
	case Of:
		return "of"
	case Fun:
		return "fun"
	case When:
		return "when"
	case Receive:
		return "receive"
	case After:
		return "after"
	case Try:
		return "try"
	case Recover:
		return "recover"
	default:
		return "unknown token"
	}
}
