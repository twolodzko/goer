package core

import (
	"fmt"
	"time"

	"github.com/twolodzko/goer/core/envir"
	"github.com/twolodzko/goer/core/errors"
	"github.com/twolodzko/goer/core/pids"
	. "github.com/twolodzko/goer/types"
)

// Setup the new, initialized environment.
func NewEnv() *envir.Env {
	return envir.InitEnv(buildIns)
}

// The build-in functions.
type buildIn = func([]Expr, *envir.Env, pids.Pid) (Expr, error)

// Initialize the build-in functions for the Env.
func buildIns() map[string]Expr {
	vars := make(map[string]Expr)
	vars["exit"] = oneArg(exit)
	vars["include"] = include
	vars["is_atom"] = oneArg(is_type[Atom])
	vars["is_bool"] = oneArg(is_type[Bool])
	vars["is_int"] = oneArg(is_type[Int])
	vars["is_list"] = oneArg(is_type[List])
	vars["is_str"] = oneArg(is_type[String])
	vars["is_tuple"] = oneArg(is_type[Tuple])
	vars["last"] = oneArg(last)
	vars["len"] = oneArg(length)
	vars["print"] = oneArg(print)
	vars["rest"] = oneArg(rest)
	vars["rev"] = oneArg(rev)
	vars["self"] = self
	vars["sleep"] = oneArg(sleep)
	vars["spawn"] = oneArg(spawn)
	vars["split"] = oneArg(split)
	vars["str"] = oneArg(str)
	return vars
}

// exit/1
func exit(arg Expr) (Expr, error) {
	return nil, errors.Exit{arg}
}

// print/1
func print(arg Expr) (Expr, error) {
	switch expr := arg.(type) {
	case String:
		fmt.Print(string(expr))
		return expr, nil
	default:
		s := fmt.Sprint(expr)
		fmt.Print(s)
		return String(s), nil
	}
}

// self/0
func self(args []Expr, env *envir.Env, pid pids.Pid) (Expr, error) {
	if len(args) > 0 {
		return nil, errors.WrongNumberArgs{}
	}
	return pid, nil
}

// sleep/1
func sleep(arg Expr) (Expr, error) {
	switch expr := arg.(type) {
	case Int:
		time.Sleep(time.Duration(expr) * time.Millisecond)
		return expr, nil
	default:
		return nil, errors.NotNumber{expr}
	}
}

// split/1
func split(arg Expr) (Expr, error) {
	str, ok := arg.(String)
	if !ok {
		return nil, errors.NotString{arg}
	}
	lst := List{}
	for _, s := range str {
		lst.Values = append(lst.Values, String(s))
	}
	return lst, nil
}

// str/1
func str(arg Expr) (Expr, error) {
	switch expr := arg.(type) {
	case String:
		return expr, nil
	default:
		return String(fmt.Sprintf("%v", expr)), nil
	}
}

// include/1
func include(args []Expr, env *envir.Env, pid pids.Pid) (Expr, error) {
	if len(args) != 1 {
		return nil, errors.WrongNumberArgs{}
	}
	path, ok := args[0].(String)
	if !ok {
		return nil, errors.NotString{args[0]}
	}
	return EvalFile(string(path), env, pid)
}

// is_T/1 generic method
func is_type[T Expr](arg Expr) (Expr, error) {
	_, ok := arg.(T)
	return Bool(ok), nil
}

// last/1
func last(arg Expr) (Expr, error) {
	switch expr := arg.(type) {
	case List:
		if expr.Len() > 0 {
			return expr.Values[expr.Len()-1], nil
		}
		return nil, errors.EmptyList{}
	default:
		return nil, errors.NotList{expr}
	}
}

// length/1
func length(arg Expr) (Expr, error) {
	switch expr := arg.(type) {
	case List:
		return Int(expr.Len()), nil
	default:
		return nil, errors.NotList{expr}
	}
}

// rest/1
func rest(arg Expr) (Expr, error) {
	switch expr := arg.(type) {
	case List:
		switch expr.Len() {
		case 0:
			return nil, errors.EmptyList{}
		case 1:
			return List{}, nil
		default:
			return List{expr.Values[:expr.Len()-1]}, nil
		}
	default:
		return nil, errors.NotList{expr}
	}
}

// rev/1
func rev(arg Expr) (Expr, error) {
	switch expr := arg.(type) {
	case List:
		n := expr.Len()
		if n == 0 {
			return List{}, nil
		}
		reversed := make([]Expr, n)
		for i := 0; i < n; i++ {
			reversed[i] = expr.Values[n-i-1]
		}
		return List{reversed}, nil
	default:
		return nil, errors.NotList{expr}
	}
}

// Decorate simple, single-argument function as a proper buildIn.
func oneArg(fun func(Expr) (Expr, error)) func([]Expr, *envir.Env, pids.Pid) (Expr, error) {
	return func(args []Expr, _ *envir.Env, _ pids.Pid) (Expr, error) {
		if len(args) != 1 {
			return nil, errors.WrongNumberArgs{}
		}
		return fun(args[0])
	}
}
