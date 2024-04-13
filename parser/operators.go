package parser

// Operator precedence (lower means higher priority),
// see: https://www.erlang.org/doc/reference_manual/expressions#operator-precedence
var operatorPrecedence = map[string]int{
	// priority 1: :
	// priority 2: #
	// priority 3: Unary + - bnot not
	"*":   4,
	"/":   4,
	"rem": 4,
	"+":   5,
	"-":   5,
	"++":  6,
	"!=":  7, // instead of "/="
	"==":  7,
	"<":   7,
	"<=":  7, // instead of "=<"
	">":   7,
	">=":  7,
	"and": 8,
	"or":  9,
	"!":   10,
	"=":   10,
	// priority 11: catch
}
