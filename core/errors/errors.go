package errors

import (
	"fmt"

	. "github.com/twolodzko/goer/types"
)

type Unbound struct{ Name string }

func (err Unbound) Error() string {
	return fmt.Sprintf("variable '%v' is unbound", err.Name)
}

type NoMatch struct{ Lhs, Rhs Expr }

func (err NoMatch) Error() string {
	return fmt.Sprintf("'%v' and '%v' do not match", err.Lhs, err.Rhs)
}

type NotNumber struct{ Value Expr }

func (err NotNumber) Error() string {
	return fmt.Sprintf("'%v' is not a number", err.Value)
}

type NotBoolean struct{ Value Expr }

func (err NotBoolean) Error() string {
	return fmt.Sprintf("'%v' is not a boolean", err.Value)
}

type NotString struct{ Value Expr }

func (err NotString) Error() string {
	return fmt.Sprintf("'%v' is not a string", err.Value)
}

type NotName struct{ Value Expr }

func (err NotName) Error() string {
	return fmt.Sprintf("'%v' is not a valid name", err.Value)
}

type NotList struct{ Value Expr }

func (err NotList) Error() string {
	return fmt.Sprintf("'%v' is not a list", err.Value)
}

type DivisionByZero struct{}

func (err DivisionByZero) Error() string {
	return "division by zero"
}

type NoTrueBranch struct{}

func (err NoTrueBranch) Error() string {
	return "no true branch found"
}

type NotFunction struct{ Value Expr }

func (err NotFunction) Error() string {
	return fmt.Sprintf("'%v' is not a function", err.Value)
}

type NoFunBranch struct{}

func (err NoFunBranch) Error() string {
	return "arguments do not match the function definition"
}

type WrongNumberArgs struct{}

func (err WrongNumberArgs) Error() string {
	return fmt.Sprintf("wrong number of arguments")
}

type EmptyList struct{}

func (err EmptyList) Error() string {
	return "empty list"
}

type Exit struct{ Reason Expr }

func (err Exit) Error() string {
	return fmt.Sprintf("exception exit: %v", err.Reason)
}

type Custom struct{ Msg string }

func (err Custom) Error() string {
	return err.Msg
}

// Create a custom error message from the format template and arguments.
func New(format string, a ...any) Custom {
	return Custom{fmt.Sprintf(format, a...)}
}
