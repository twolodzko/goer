package core

import (
	"math"
	"time"

	"github.com/twolodzko/goer/core/envir"
	"github.com/twolodzko/goer/core/errors"
	"github.com/twolodzko/goer/core/pids"
	. "github.com/twolodzko/goer/types"
)

const defaultTimeout time.Duration = math.MaxInt64

// Send `msg` message to the channel with pid `to`.
func send(to, msg Expr) (Expr, error) {
	switch pid := to.(type) {
	case pids.Pid:
		pid.Send(msg)
		return msg, nil
	default:
		return nil, errors.New("%v is not a pid", to)
	}
}

// spawn/1
func spawn(arg Expr) (Expr, error) {
	fun, ok := arg.(Fun)
	if !ok {
		return nil, errors.NotFunction{arg}
	}
	pid := pids.NewPid()

	go func() {
		defer pid.Close()
		expr, env, err := fun.call(nil, pid)
		if err == nil {
			Eval(expr, env, pid)
		}
	}()

	return pid, nil
}

// Run the receive block.
func receive(receive Receive, env *envir.Env, pid pids.Pid) (Expr, *envir.Env, error) {
	timeout, err := getTimeout(receive, env, pid)
	if err != nil {
		return nil, env, err
	}

	// loop is needed since we ignore messages that
	// do not fit the patterns
	for {
		select {
		case msg := <-pid.Messages():
			for _, branch := range receive.Branches {
				if match(branch.Pattern, msg, env, pid) == nil {
					ok, err := evalAllTrue(branch.Guards, env, pid)
					if ok && err == nil {
						return partialEval(branch.Body, env, pid)
					}
				}
			}
		case <-time.After(timeout):
			return partialEval(receive.After.Body, env, pid)
		}
	}
}

// Get the timeout value for receive.
func getTimeout(r Receive, env *envir.Env, pid pids.Pid) (time.Duration, error) {
	if r.After.Cond != nil {
		expr, err := Eval(r.After.Cond, env, pid)
		if err != nil {
			return 0, err
		}

		switch t := expr.(type) {
		case Int:
			return time.Duration(t) * time.Millisecond, nil
		case Atom:
			if t == "infinity" {
				return defaultTimeout, nil
			}
		}
		return 0, errors.NotNumber{expr}
	}
	return defaultTimeout, nil
}
