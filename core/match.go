package core

import (
	"reflect"

	"github.com/twolodzko/goer/core/envir"
	"github.com/twolodzko/goer/core/errors"
	"github.com/twolodzko/goer/core/pids"
	. "github.com/twolodzko/goer/types"
)

// The Erlang's match (=) operation.
func match(lhs, rhs Expr, env *envir.Env, pid pids.Pid) error {
	var (
		final bool
		err   error
	)
	// treat lhs as a key and match, otherwise evaluate it
	lhs, final, err = evalMatch(lhs, rhs, env, pid)
	if err != nil || final {
		return err
	}
	// or the other way around
	rhs, final, err = evalMatch(rhs, lhs, env, pid)
	if err != nil || final {
		return err
	}

	if reflect.TypeOf(lhs) == reflect.TypeOf(rhs) {
		switch lhs := lhs.(type) {
		case List:
			rhs := rhs.(List)
			return matchAll(lhs.Values, rhs.Values, env, pid)
		case Tuple:
			rhs := rhs.(Tuple)
			return matchAll(lhs.Values, rhs.Values, env, pid)
		default:
			if lhs == rhs {
				return nil
			}
		}
	}

	return errors.NoMatch{lhs, rhs}
}

// Try matching value against key, otherwise evaluate the key.
// If the key is a container (List or Tuple), you need to handle it separately.
func evalMatch(key, val Expr, env *envir.Env, pid pids.Pid) (Expr, bool, error) {
	var err error
	switch name := key.(type) {
	case Dummy:
		return key, true, nil
	case Variable:
		if _, ok := val.(Dummy); ok {
			return key, true, nil
		}
		rhs, err := Eval(val, env, pid)
		if err != nil {
			return key, true, err
		}
		return key, true, env.TrySet(name, rhs)
	case List, Tuple:
		// handle recursive case separately
	default:
		key, err = Eval(key, env, pid)
	}
	return key, false, err
}

// Apply match to the elements of slices.
func matchAll(lhs, rhs []Expr, env *envir.Env, pid pids.Pid) error {
	if len(lhs) != len(rhs) {
		return errors.NoMatch{lhs, rhs}
	}
	for i := range lhs {
		err := match(lhs[i], rhs[i], env, pid)
		if err != nil {
			return err
		}
	}
	return nil
}
