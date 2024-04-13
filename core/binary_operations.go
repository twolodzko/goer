package core

import (
	"fmt"
	"reflect"

	"github.com/twolodzko/goer/core/errors"
	. "github.com/twolodzko/goer/types"
)

func applyBinaryOp(op string, lhs, rhs Expr) (Expr, error) {
	switch op {
	case "and":
		lhs, rhs, err := maybeBools(lhs, rhs)
		return lhs && rhs, err
	case "or":
		lhs, rhs, err := maybeBools(lhs, rhs)
		return lhs || rhs, err
	case "+":
		lhs, rhs, err := maybeInts(lhs, rhs)
		return lhs + rhs, err
	case "-":
		lhs, rhs, err := maybeInts(lhs, rhs)
		return lhs - rhs, err
	case "*":
		lhs, rhs, err := maybeInts(lhs, rhs)
		return lhs * rhs, err
	case "/":
		lhs, rhs, err := maybeInts(lhs, rhs)
		if rhs == 0 {
			return nil, errors.DivisionByZero{}
		}
		return lhs / rhs, err
	case "rem":
		lhs, rhs, err := maybeInts(lhs, rhs)
		if rhs == 0 {
			return nil, errors.DivisionByZero{}
		}
		return lhs % rhs, err
	case "<":
		lhs, rhs, err := maybeInts(lhs, rhs)
		return Bool(lhs < rhs), err
	case "<=":
		lhs, rhs, err := maybeInts(lhs, rhs)
		return Bool(lhs <= rhs), err
	case ">":
		lhs, rhs, err := maybeInts(lhs, rhs)
		return Bool(lhs > rhs), err
	case ">=":
		lhs, rhs, err := maybeInts(lhs, rhs)
		return Bool(lhs >= rhs), err
	case "==":
		return Bool(reflect.DeepEqual(rhs, lhs)), nil
	case "!=":
		return Bool(!reflect.DeepEqual(rhs, lhs)), nil
	case "++":
		switch lhs := lhs.(type) {
		case List:
			return listAppend(lhs, rhs)
		case String:
			return stringAppend(lhs, rhs)
		default:
			return nil, errors.New("unexpected value: %v", lhs)
		}
	case "!":
		return send(lhs, rhs)
	default:
		panic(fmt.Sprintf("invalid operator '%s'", op))
	}
}

// Concatenate two lists or add an element to the list.
func listAppend(lhs List, rhs Expr) (Expr, error) {
	switch rhs := rhs.(type) {
	case List:
		return lhs.Append(rhs.Values...), nil
	default:
		return lhs.Append(rhs), nil
	}
}

// Concatenate two strings.
func stringAppend(lhs String, rhs Expr) (Expr, error) {
	switch rhs := rhs.(type) {
	case String:
		return lhs + rhs, nil
	default:
		return nil, errors.NotString{rhs}
	}
}

// Try casting the expressions to integers.
func maybeInts(lhs, rhs Expr) (Int, Int, error) {
	x, err := maybeInt(lhs)
	if err != nil {
		return 0, 0, err
	}
	y, err := maybeInt(rhs)
	return x, y, err
}

// Try casting the expressions to booleans.
func maybeBools(lhs, rhs Expr) (Bool, Bool, error) {
	switch lhs := lhs.(type) {
	case Bool:
		switch rhs := rhs.(type) {
		case Bool:
			return lhs, rhs, nil
		default:
			return false, false, errors.NotBoolean{rhs}
		}
	default:
		return false, false, errors.NotBoolean{lhs}
	}
}
