package envir

import (
	"reflect"

	"github.com/twolodzko/goer/core/errors"
	. "github.com/twolodzko/goer/types"
)

type Env struct {
	Elems  map[string]Expr
	parent *Env
}

// Create an uninitialized Env (mostly for testing).
func EmptyEnv() *Env {
	return InitEnv(func() map[string]Expr { return make(map[string]Expr) })
}

// Create Env, use the `init` functions to initialize it.
func InitEnv(init func() map[string]Expr) *Env {
	vars := init()
	return &Env{vars, nil}
}

// Create child Env from the parent (current).
func (parent *Env) Branch() *Env {
	vars := make(map[string]Expr)
	return &Env{vars, parent}
}

// Get the value from Env, error if not available.
func (env *Env) Get(key Expr) (Expr, error) {
	name, err := getName(key)
	if err != nil {
		return nil, err
	}

	current := env
	for current != nil {
		if val, ok := current.Elems[name]; ok {
			return val, nil
		}
		current = current.parent
	}
	return nil, errors.Unbound{name}
}

// Try to set the value in the Env.
// If the value exists and differs from the given argument, throw an error.
func (env *Env) TrySet(key, value Expr) error {
	name, err := getName(key)
	if err != nil {
		return err
	}

	// if it exists
	if prev, ok := env.Elems[name]; ok {
		if !reflect.DeepEqual(prev, value) {
			return errors.NoMatch{key, value}
		}
		return nil
	}

	env.Elems[name] = value
	return nil
}

func getName(key Expr) (string, error) {
	switch key := key.(type) {
	case Variable:
		return string(key), nil
	case Atom:
		return string(key), nil
	default:
		return "", errors.NotName{key}
	}
}
