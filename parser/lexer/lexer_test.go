package lexer

import (
	"testing"
	"unicode"

	"github.com/google/go-cmp/cmp"
)

func TestTokenize(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		input    string
		expected []Token
	}{
		{"", nil},
		{"foo", []Token{{Atom, "foo"}}},
		{"bar   ", []Token{{Atom, "bar"}}},
		{"other_@Atom", []Token{{Atom, "other_@Atom"}}},
		{"X", []Token{{Variable, "X"}}},
		{" Abc	", []Token{{Variable, "Abc"}}},
		{" _ ", []Token{{Dummy, "_"}}},
		{" _This	", []Token{{Variable, "_This"}}},
		{"42", []Token{{Number, "42"}}},
		{"  123  ", []Token{{Number, "123"}}},
		{".", []Token{{Dot, "."}}},
		{",", []Token{{Comma, ","}}},
		{";", []Token{{Semicolon, ";"}}},
		{"()", []Token{{BracketLeft, "("}, {BracketRight, ")"}}},
		{"{}", []Token{{BraceLeft, "{"}, {BraceRight, "}"}}},
		{"+", []Token{{Operator, "+"}}},
		{"->", []Token{{Arrow, "->"}}},
		{"when", []Token{{When, "when"}}},
		{"not Thing", []Token{{Operator, "not"}, {Variable, "Thing"}}},
		{"2 + 3", []Token{{Number, "2"}, {Operator, "+"}, {Number, "3"}}},
		{"36-X", []Token{{Number, "36"}, {Operator, "-"}, {Variable, "X"}}},
		{"8*14", []Token{{Number, "8"}, {Operator, "*"}, {Number, "14"}}},
		{"<=", []Token{{Operator, "<="}}},
		{"21=/=12", []Token{{Number, "21"}, {Operator, "=/="}, {Number, "12"}}},
		{"2*PI", []Token{{Number, "2"}, {Operator, "*"}, {Variable, "PI"}}},
		{"(2+7)/3", []Token{
			{BracketLeft, "("}, {Number, "2"}, {Operator, "+"}, {Number, "7"}, {BracketRight, ")"},
			{Operator, "/"}, {Number, "3"},
		}},
		{"% hey, skip this comment\n ok", []Token{{Atom, "ok"}}},
		{"alone  % also skip that comment", []Token{{Atom, "alone"}}},
		{"(_)", []Token{{BracketLeft, "("}, {Dummy, "_"}, {BracketRight, ")"}}},
		{`""`, []Token{{String, `""`}}},
		{`"Hello, World!"`, []Token{{String, `"Hello, World!"`}}},
		{`"\"Hello,\nWorld!\""`, []Token{{String, `"\"Hello,\nWorld!\""`}}},
	}

	for _, tt := range testCases {
		result, err := Tokenize(tt.input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !cmp.Equal(result, tt.expected) {
			t.Errorf("for '%q' expected %v, got: %v", tt.input, tt.expected, result)
		}
	}
}

func TestTakeWhile(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		input string
		fun   func(rune) bool
		pos   int
	}{
		{"", unicode.IsSpace, 0},
		{" one", unicode.IsSpace, 1},
		{"   three", unicode.IsSpace, 3},
		{" \t\t  mixed", unicode.IsSpace, 5},
		{"aaabcdaa", func(r rune) bool { return r == 'a' }, 3},
		{"aaabcdaa", func(r rune) bool { return r == 'Z' }, 1},
	}

	for _, tt := range testCases {
		lx := newLexer(tt.input)
		lx.next()
		lx.takeWhile(tt.fun)
		if lx.pos != tt.pos {
			t.Errorf("for '%q' expected %v, got: %v", tt.input, tt.pos, lx.pos)
		}
	}
}
