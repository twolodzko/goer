package core

import (
	"fmt"

	"github.com/twolodzko/goer/core/envir"
	"github.com/twolodzko/goer/core/errors"
	"github.com/twolodzko/goer/core/pids"
	. "github.com/twolodzko/goer/types"
)

// A function that encloses the environment where it was defined.
type Fun struct {
	parentEnv *envir.Env
	Definition
}

// Call the function with the arguments.
func (fun Fun) call(args []Expr, pid pids.Pid) (Expr, *envir.Env, error) {
	for _, branch := range fun.Branches {
		env := fun.parentEnv.Branch()
		if matchAll(branch.Args, args, env, pid) == nil {
			ok, err := evalAllTrue(branch.Guards, env, pid)
			if err != nil {
				return nil, env, err
			}
			if ok {
				return partialEval(branch.Body, env, pid)
			}
		}
	}
	return nil, fun.parentEnv, errors.NoFunBranch{}
}

func (fun Fun) String() string {
	return fmt.Sprintf("%s", fun.Definition)
}

// Evaluate an if expression.
func evalIf(block If, env *envir.Env, pid pids.Pid) (Expr, *envir.Env, error) {
	for _, branch := range block.Branches {
		// directly handle booleans and dummies
		if isTrueish(branch.Cond) {
			return partialEval(branch.Body, env, pid)
		}

		cond, err := evalIsTrue(branch.Cond, env, pid)
		if err != nil {
			return nil, env, err
		}
		if cond {
			return partialEval(branch.Body, env, pid)
		}
	}
	return nil, env, errors.NoTrueBranch{}
}

func evalCase(block Case, env *envir.Env, pid pids.Pid) (Expr, *envir.Env, error) {
	val, err := Eval(block.Arg, env, pid)
	if err != nil {
		return nil, env, err
	}
	for _, branch := range block.Branches {
		// no match error = true
		if match(val, branch.Pattern, env, pid) == nil {
			ok, err := evalAllTrue(branch.Guards, env, pid)
			if err != nil {
				return nil, env, err
			}
			if ok {
				return partialEval(branch.Body, env, pid)
			}
		}
	}
	return nil, env, errors.NoTrueBranch{}
}

// Is the expression true-ish (bool or dummy).
func isTrueish(expr Expr) bool {
	switch val := expr.(type) {
	case Dummy:
		return true
	case Bool:
		return bool(val)
	default:
		return false
	}
}

// Evaluate all expressions (short circuit), check if they are all true.
func evalAllTrue(exprs []Expr, env *envir.Env, pid pids.Pid) (Bool, error) {
	for _, expr := range exprs {
		ok, err := evalIsTrue(expr, env, pid)
		if !ok || err != nil {
			return false, err
		}
	}
	return true, nil
}

// Evaluate expression and check if it is true.
func evalIsTrue(expr Expr, env *envir.Env, pid pids.Pid) (Bool, error) {
	expr, err := Eval(expr, env, pid)
	if err != nil {
		return false, err
	}
	switch expr := expr.(type) {
	case Bool:
		return expr, nil
	default:
		return false, errors.NotBoolean{expr}
	}
}
