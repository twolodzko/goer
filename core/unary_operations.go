package core

import (
	"fmt"

	"github.com/twolodzko/goer/core/errors"
	. "github.com/twolodzko/goer/types"
)

func applyUnaryOp(op string, expr Expr) (Expr, error) {
	switch op {
	case "+":
		return maybeInt(expr)
	case "-":
		val, err := maybeInt(expr)
		return -val, err
	case "not":
		switch expr := expr.(type) {
		case Bool:
			return !expr, nil
		default:
			return nil, errors.NotBoolean{expr}
		}
	default:
		panic(fmt.Sprintf("invalid operator '%s'", op))
	}
}

// Try casting the expression to an integer.
func maybeInt(expr Expr) (Int, error) {
	switch expr := expr.(type) {
	case Int:
		return expr, nil
	default:
		return 0, errors.NotNumber{expr}
	}
}
