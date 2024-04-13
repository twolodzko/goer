package envir

import (
	"testing"

	"github.com/twolodzko/goer/core/errors"
	. "github.com/twolodzko/goer/types"
)

func TestSetGet(t *testing.T) {
	t.Parallel()

	var (
		err error
		val Expr
	)

	parent := EmptyEnv()

	// get X
	val, err = parent.Get(Variable("X"))
	if err == nil {
		t.Errorf("querying for X in empty environment should result in an error")
		return
	}

	// X = 1
	err = parent.TrySet(Variable("X"), Int(1))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	val, err = parent.Get(Variable("X"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if val != Int(1) {
		t.Errorf("when querying child for X we got %v", val)
	}
	// Y = 2
	err = parent.TrySet(Variable("Y"), Int(2))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	val, err = parent.Get(Variable("Y"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if val != Int(2) {
		t.Errorf("when querying child for Y we got %v", val)
	}

	child := parent.Branch()

	// Y = 3
	err = child.TrySet(Variable("Y"), Int(3))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	val, err = child.Get(Variable("Y"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if val != Int(3) {
		t.Errorf("when querying child for Y we got %v", val)
	}

	// Z = 4
	err = child.TrySet(Variable("Z"), Int(4))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	val, err = child.Get(Variable("Z"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if val != Int(4) {
		t.Errorf("when querying child for Z we got %v", val)
	}

	// get X from child
	val, err = child.Get(Variable("X"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if val != Int(1) {
		t.Errorf("when querying child for X we got %v", val)
	}

	// get Z from parent
	val, err = parent.Get(Variable("Z"))
	if (err != errors.Unbound{"Z"}) {
		t.Errorf("when querying parent for Z we got %s instead of error", err)
		return
	}

}

func TestErrors(t *testing.T) {
	env := EmptyEnv()
	if err := env.TrySet(Int(1), Int(1)); err == nil {
		t.Error("set didn't error")
	}
	if _, err := env.Get(Int(1)); err == nil {
		t.Error("set didn't error")
	}

	if err := env.TrySet(Variable("X"), Int(1)); err != nil {
		panic(err)
	}
	if err := env.TrySet(Variable("X"), Int(2)); err == nil {
		t.Error("set didn't error")
	}
}
