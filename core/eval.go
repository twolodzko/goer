package core

import (
	"fmt"
	"io"
	"os"

	"github.com/twolodzko/goer/core/envir"
	"github.com/twolodzko/goer/core/errors"
	"github.com/twolodzko/goer/core/pids"
	"github.com/twolodzko/goer/parser"
	"github.com/twolodzko/goer/parser/reader"
	. "github.com/twolodzko/goer/types"
)

// Evaluate an expression.
func Eval(expr Expr, env *envir.Env, pid pids.Pid) (Expr, error) {
	var err error
	for {
		switch val := expr.(type) {
		case Variable:
			return env.Get(val)
		case Dummy:
			return nil, errors.Unbound{"_"}
		case Atom:
			if val, err := env.Get(val); err == nil {
				return val, nil
			}
			return val, nil
		case Bool, Int, String, pids.Pid, Fun:
			return val, nil
		case Tuple:
			exprs, err := evalAll(val.Values, env, pid)
			return Tuple{exprs}, err
		case List:
			exprs, err := evalAll(val.Values, env, pid)
			return List{exprs}, err
		case UnaryOperation:
			rhs, err := Eval(val.Rhs, env, pid)
			if err != nil {
				return nil, err
			}
			return applyUnaryOp(val.Op, rhs)
		case BinaryOperation:
			switch val.Op {
			case "=":
				err := match(val.Lhs, val.Rhs, env, pid)
				return Bool(err == nil), err
			default:
				lhs, err := Eval(val.Lhs, env, pid)
				if err != nil {
					return nil, err
				}
				rhs, err := Eval(val.Rhs, env, pid)
				if err != nil {
					return nil, err
				}
				return applyBinaryOp(val.Op, lhs, rhs)
			}
		case Bracket:
			expr = val.Expr
		case If:
			expr, env, err = evalIf(val, env, pid)
			if err != nil {
				return nil, err
			}
		case Case:
			expr, env, err = evalCase(val, env, pid)
			if err != nil {
				return nil, err
			}
		case TryRecover:
			if expr, err := EvalBlock(val.Body, env, pid); err == nil {
				return expr, nil
			} else {
				return EvalBlock(val.Recover, env, pid)
			}
		case Definition:
			fun := Fun{env, val}
			if val.Name != "" {
				name := string(val.Name)
				_, exists := env.Elems[name]
				if exists {
					return nil, errors.New("%s already exists", name)
				}
				env.Elems[name] = fun
			}
			return fun, err
		case Call:
			args, err := evalAll(val.Args, env, pid)
			if err != nil {
				return nil, err
			}
			fun, err := Eval(val.Callable, env, pid)
			if err != nil {
				return nil, err
			}

			switch fun := fun.(type) {
			case Fun:
				expr, env, err = fun.call(args, pid)
				if err != nil {
					return nil, err
				}
			case buildIn:
				return fun(args, env, pid)
			default:
				return nil, errors.NotFunction{val.Callable}
			}
		case Receive:
			expr, env, err = receive(val, env, pid)
			if err != nil {
				return nil, err
			}
		default:
			panic(fmt.Sprintf("value of type %T cannot be evaluated", val))
		}
	}
}

// Evaluate a block of code, return result of the last expression.
func EvalBlock(exprs []Expr, env *envir.Env, pid pids.Pid) (Expr, error) {
	expr, env, err := partialEval(exprs, env, pid)
	if err != nil {
		return nil, err
	}
	return Eval(expr, env, pid)
}

// Evaluate list of expressions.
func evalAll(exprs []Expr, env *envir.Env, pid pids.Pid) ([]Expr, error) {
	var evaluated []Expr
	for _, expr := range exprs {
		val, err := Eval(expr, env, pid)
		if err != nil {
			return evaluated, err
		}
		evaluated = append(evaluated, val)
	}
	return evaluated, nil
}

// Evaluate list of expressions, return last expression not evaluated.
func partialEval(exprs []Expr, env *envir.Env, pid pids.Pid) (Expr, *envir.Env, error) {
	n := len(exprs)
	if n == 0 {
		return nil, env, nil
	}
	for i := 0; i < n-1; i++ {
		_, err := Eval(exprs[i], env, pid)
		if err != nil {
			return nil, env, err
		}
	}
	return exprs[n-1], env, nil
}

// Parse the code string and evaluate it.
func ParseEval(code string, env *envir.Env, pid pids.Pid) (Expr, error) {
	exprs, err := parser.Parse(code)
	if err != nil {
		return nil, err
	}
	return EvalBlock(exprs, env, pid)
}

// Evaluate a file.
func EvalFile(path string, env *envir.Env, pid pids.Pid) (Expr, error) {
	var expr Expr = Atom("ok")

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := reader.NewReader(file)
	for {
		code, err := reader.Next()
		switch err {
		case nil:
		case io.EOF:
			return expr, nil
		default:
			return nil, err
		}

		expr, err = ParseEval(code, env, pid)
		if err != nil {
			return nil, err
		}
	}
}
