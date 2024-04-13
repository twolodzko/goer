package parser

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/twolodzko/goer/parser/lexer"
	. "github.com/twolodzko/goer/types"
)

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected []Expr
	}{
		// basic data types
		{"1 .", []Expr{Int(1)}},
		{"foo .", []Expr{Atom("foo")}},
		{"X .", []Expr{Variable("X")}},
		{"true .", []Expr{Bool(true)}},
		{"false .", []Expr{Bool(false)}},
		{`"".`, []Expr{String("")}},
		{`"Hello, World!".`, []Expr{String("Hello, World!")}},
		{`"\"Hello,\nWorld!\"".`, []Expr{String("\"Hello,\nWorld!\"")}},
		{"{}.", []Expr{Tuple{}}},
		{"{1,2,3}.", []Expr{Tuple{[]Expr{Int(1), Int(2), Int(3)}}}},
		{"[].", []Expr{List{}}},
		{"[1,2,3].", []Expr{List{[]Expr{Int(1), Int(2), Int(3)}}}},

		// series of expressions
		{"1,2,3.", []Expr{Int(1), Int(2), Int(3)}},

		// unary operations
		{"- 5 .", []Expr{UnaryOperation{"-", Int(5)}}},
		{"+ Y .", []Expr{UnaryOperation{"+", Variable("Y")}}},
		{"not X .", []Expr{UnaryOperation{"not", Variable("X")}}},

		// binary expressions
		{"X = 42.", []Expr{BinaryOperation{"=", Variable("X"), Int(42)}}},
		{"2 + 3.", []Expr{BinaryOperation{"+", Int(2), Int(3)}}},
		{"2 + 3 - 4 + 5.", []Expr{
			BinaryOperation{"+",
				BinaryOperation{"-",
					BinaryOperation{"+", Int(2), Int(3)},
					Int(4),
				},
				Int(5),
			},
		}},
		{"6 + 7 * 8 / 9.", []Expr{
			BinaryOperation{"+",
				Int(6),
				BinaryOperation{"/",
					BinaryOperation{"*", Int(7), Int(8)},
					Int(9),
				},
			},
		}},
		{"(2 + 2).", []Expr{Bracket{BinaryOperation{"+", Int(2), Int(2)}}}},
		{"(2 + 4) / 3.", []Expr{
			BinaryOperation{"/",
				Bracket{BinaryOperation{"+", Int(2), Int(4)}},
				Int(3),
			},
		}},
		{"(2 + 4) / (3 - 5).", []Expr{
			BinaryOperation{"/",
				Bracket{BinaryOperation{"+", Int(2), Int(4)}},
				Bracket{BinaryOperation{"-", Int(3), Int(5)}},
			},
		}},
		{"(1+2)/3.", []Expr{
			BinaryOperation{"/",
				Bracket{BinaryOperation{"+", Int(1), Int(2)}},
				Int(3),
			},
		}},
		{"X = 1/2 + 3.", []Expr{
			BinaryOperation{"=",
				Variable("X"),
				BinaryOperation{"+",
					BinaryOperation{"/", Int(1), Int(2)},
					Int(3),
				},
			},
		}},
		{"_ = true.", []Expr{
			BinaryOperation{"=",
				Dummy{},
				Bool(true),
			},
		}},

		// functions
		{"foo().", []Expr{Call{Atom("foo"), nil}}},
		{"Bar().", []Expr{Call{Variable("Bar"), nil}}},
		{"identity(X).", []Expr{Call{Atom("identity"), []Expr{Variable("X")}}}},
		{"Identity(X).", []Expr{Call{Variable("Identity"), []Expr{Variable("X")}}}},
		{"fun(X) -> X end.", []Expr{
			Definition{
				"",
				[]FunBranch{
					{
						[]Expr{Variable("X")},
						nil,
						[]Expr{Variable("X")},
					},
				}}},
		},
		{"fun (0) -> true; (_) -> false end.", []Expr{
			Definition{
				"",
				[]FunBranch{
					{[]Expr{Int(0)}, nil, []Expr{Bool(true)}},
					{[]Expr{Dummy{}}, nil, []Expr{Bool(false)}},
				}},
		}},
		{"(fun(X) -> X end)(true).", []Expr{
			Call{
				Bracket{
					Definition{
						"",
						[]FunBranch{
							{
								[]Expr{Variable("X")},
								nil,
								[]Expr{Variable("X")},
							},
						}}},
				[]Expr{Bool(true)},
			},
		}},
		{"fun identity(X) -> X end.", []Expr{
			Definition{
				"identity",
				[]FunBranch{
					{
						[]Expr{Variable("X")},
						nil,
						[]Expr{Variable("X")},
					},
				}}},
		},

		// control flow
		{"if X == true -> true end.", []Expr{If{
			[]IfBranch{
				{
					BinaryOperation{
						"==",
						Variable("X"),
						Bool(true),
					},
					[]Expr{Bool(true)},
				},
			},
		}}},
		{"if X == 1 -> true; _ -> false end.", []Expr{
			If{[]IfBranch{
				{
					BinaryOperation{
						"==",
						Variable("X"),
						Int(1),
					},
					[]Expr{Bool(true)},
				},
				{
					Dummy{},
					[]Expr{Bool(false)},
				},
			}},
		}},
		{"case X of Y when Y == 1 -> Y+2 end.", []Expr{Case{
			Variable("X"),
			[]PatternBranch{
				{
					Variable("Y"),
					[]Expr{BinaryOperation{"==", Variable("Y"), Int(1)}},
					[]Expr{BinaryOperation{"+", Variable("Y"), Int(2)}},
				},
			},
		}}},
		{"case X of true -> 1; false -> 2 end.", []Expr{Case{
			Variable("X"),
			[]PatternBranch{
				{
					Bool(true),
					nil,
					[]Expr{Int(1)},
				},
				{
					Bool(false),
					nil,
					[]Expr{Int(2)},
				},
			},
		}}},
		{"try 1/0 recover nan end.", []Expr{
			TryRecover{
				[]Expr{BinaryOperation{"/", Int(1), Int(0)}},
				[]Expr{Atom("nan")},
			},
		}},

		// receive
		{"receive true -> 1; false -> 2 end.", []Expr{Receive{
			[]PatternBranch{
				{Bool(true), nil, []Expr{Int(1)}},
				{Bool(false), nil, []Expr{Int(2)}},
			},
			IfBranch{},
		}}},
		{"receive after 0 -> true end.", []Expr{Receive{
			nil,
			IfBranch{Int(0), []Expr{Bool(true)}},
		}}},
		{"receive true -> 1 after 5 -> 2 end.", []Expr{Receive{
			[]PatternBranch{
				{Bool(true), nil, []Expr{Int(1)}},
			},
			IfBranch{Int(5), []Expr{Int(2)}},
		}}},

		// full expressions
		{"- 2 + 1.", []Expr{
			BinaryOperation{"+",
				UnaryOperation{"-", Int(2)},
				Int(1),
			},
		}},
		{"2 + - 1.", []Expr{
			BinaryOperation{"+",
				Int(2),
				UnaryOperation{"-", Int(1)},
			},
		}},
		{
			"Identity = fun(X) -> X end, X = Identity(1) + 2, print(X).",
			[]Expr{
				BinaryOperation{
					"=",
					Variable("Identity"),
					Definition{
						"",
						[]FunBranch{
							{
								[]Expr{Variable("X")},
								nil,
								[]Expr{Variable("X")},
							},
						}},
				},
				BinaryOperation{
					"=",
					Variable("X"),
					BinaryOperation{
						"+",
						Call{
							Variable("Identity"),
							[]Expr{Int(1)},
						},
						Int(2),
					},
				},
				Call{
					Atom("print"),
					[]Expr{Variable("X")},
				},
			},
		},
	}
	for _, tt := range testCases {
		result, err := Parse(tt.input)
		if err != nil {
			t.Errorf("parsing %q resulted in an unexpected error: %s", tt.input, err)
		} else if reflect.TypeOf(result) != reflect.TypeOf(tt.expected) {
			t.Errorf("for %q types mismatch: %T != %T", tt.input, tt.expected, result)
		} else if !cmp.Equal(result, tt.expected) {
			t.Errorf("for %q expected %v, got: %v", tt.input, tt.expected, result)
		}
	}
}

func TestParseExceptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected error
	}{
		{"35", Missing{lexer.Dot}},
		{"(X.", Unexpected{lexer.Token{lexer.Dot, "."}}},
		{"+.", Unexpected{lexer.Token{lexer.Dot, "."}}},
		{"* 5.", Unexpected{lexer.Token{lexer.Operator, "*"}}},
		{"{1,2,3.", Unexpected{lexer.Token{lexer.Dot, "."}}},
		{"5 >+< 7 .", Unexpected{lexer.Token{lexer.Operator, ">+<"}}},
		{"1,2} .", Unexpected{lexer.Token{lexer.BraceRight, "}"}}},
		{"2 + 2) .", Unexpected{lexer.Token{lexer.BracketRight, ")"}}},
		{"1(2) .", Unexpected{lexer.Token{lexer.BracketLeft, "("}}},
		{"{1,2,3} (4,5) .", Unexpected{lexer.Token{lexer.BracketLeft, "("}}},
		{"if 1 -> 1 after 2 -> 2 end.", Unexpected{lexer.Token{lexer.After, "after"}}},
		{"case 1 of 1 -> 1 after 2 -> 2 end.", Unexpected{lexer.Token{lexer.After, "after"}}},
		{"if true -> end.", Unexpected{lexer.Token{lexer.End, "end"}}},
		{"if false -> ; true -> end.", Unexpected{lexer.Token{lexer.Semicolon, ";"}}},
		{"case X of 1 -> end.", Unexpected{lexer.Token{lexer.End, "end"}}},
		{"case X of 1 -> ; 2 -> 2 end.", Unexpected{lexer.Token{lexer.Semicolon, ";"}}},
		{"case X of 1 -> 1; 2 -> end.", Unexpected{lexer.Token{lexer.End, "end"}}},
		{"receive X -> end.", Unexpected{lexer.Token{lexer.End, "end"}}},
		{"receive X -> X; _ -> end.", Unexpected{lexer.Token{lexer.End, "end"}}},
		{"fun() -> end.", Unexpected{lexer.Token{lexer.End, "end"}}},
		{"fun (1) -> 1; (X) -> end.", Unexpected{lexer.Token{lexer.End, "end"}}},
		{"try 1/0 recover end.", Unexpected{lexer.Token{lexer.End, "end"}}},
		{"try recover ok end.", Unexpected{lexer.Token{lexer.Recover, "recover"}}},
	}
	for _, tt := range testCases {
		_, err := Parse(tt.input)
		if err == nil {
			t.Errorf("for %q no error was returned", tt.input)
		} else if !cmp.Equal(err, tt.expected) {
			t.Errorf("for %q expected error %q, got: %q", tt.input, tt.expected, err)
		}
	}
}
