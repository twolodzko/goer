package main_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/twolodzko/goer/core"
	"github.com/twolodzko/goer/core/pids"
	"github.com/twolodzko/goer/types"
)

func TestFizzBuzz(t *testing.T) {
	t.Parallel()

	env := core.NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	_, err := core.ParseEval(`
	fun fizzbuzz (X) ->
		case { X rem 3, X rem 5 } of
			{0, 0} -> fizz_buzz;
			{0, _} -> fizz;
			{_, 0} -> buzz;
			_ -> X
		end
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	testCases := []struct {
		input    types.Int
		expected types.Expr
	}{
		{1, types.Int(1)},
		{2, types.Int(2)},
		{3, types.Atom("fizz")},
		{4, types.Int(4)},
		{5, types.Atom("buzz")},
		{6, types.Atom("fizz")},
		{7, types.Int(7)},
		{8, types.Int(8)},
		{9, types.Atom("fizz")},
		{10, types.Atom("buzz")},
		{15, types.Atom("fizz_buzz")},
	}
	for _, tt := range testCases {
		result, err := core.ParseEval(fmt.Sprintf("fizzbuzz(%d).", tt.input), env, pid)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		} else if !cmp.Equal(result, tt.expected) {
			t.Errorf("for %d expected %d, got %d", tt.input, tt.expected, result)
		}
	}
}

func TestMap(t *testing.T) {
	t.Parallel()

	var (
		err              error
		result, expected types.Expr
	)
	env := core.NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	_, err = core.ParseEval(`
	fun map
		(Lst, Fun) ->
			map(Lst, Fun, []);
		([], _, Acc) ->
			rev(Acc);
		(Lst, Fun, Acc) ->
			X = last(Lst),
			map(rest(Lst), Fun, Acc ++ [Fun(X)])
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	expected = types.List{[]types.Expr{types.Int(11), types.Int(12), types.Int(13)}}
	result, err = core.ParseEval("map([1,2,3], fun(X) -> X+10 end).", env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestStateful(t *testing.T) {
	t.Parallel()

	var (
		err              error
		result, expected types.Expr
	)
	env := core.NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	_, err = core.ParseEval(`
	fun loop(State) ->
		receive
			{From, set, Newstate} ->
				From ! ok,
				loop(Newstate);
			{From, get} ->
				From ! {ok, State},
				loop(State);
			{From, bye} ->
				From ! ciao
		end
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	_, err = core.ParseEval("Pid = spawn(fun() -> loop(empty) end).", env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// read initial value
	result, err = core.ParseEval(`
	Pid ! {self(), get},
	receive
		Msg1 -> Msg1
	after
		100 -> timeout
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected = types.Tuple{[]types.Expr{types.Atom("ok"), types.Atom("empty")}}
	if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// set new value
	result, err = core.ParseEval(`
	Pid ! {self(), set, hello},
	receive
		Msg2 -> Msg2
	after
		100 -> timeout
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected = types.Atom("ok")
	if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// read new value
	result, err = core.ParseEval(`
	Pid ! {self(), get},
	receive
		Msg3 -> Msg3
	after
		100 -> timeout
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected = types.Tuple{[]types.Expr{types.Atom("ok"), types.Atom("hello")}}
	if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// set and read another value
	result, err = core.ParseEval(`
	Pid ! {self(), set, different},
	receive
		_ -> ok
	after
		100 -> timeout
	end,

	Pid ! {self(), get},
	receive
		Msg4 -> Msg4
	after
		100 -> timeout
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected = types.Tuple{[]types.Expr{types.Atom("ok"), types.Atom("different")}}
	if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// close channel
	result, err = core.ParseEval(`
	Pid ! {self(), bye},
	receive
		Msg5 -> Msg5
	after
		100 -> timeout
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected = types.Atom("ciao")
	if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// it should time out because it's closed
	result, err = core.ParseEval(`
	Pid ! {self(), get},
	receive
		Msg6 -> Msg6
	after
		100 -> timeout
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected = types.Atom("timeout")
	if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func BenchmarkFact(b *testing.B) {

	env := core.NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	_, err := core.ParseEval(`
	fun fact
		(0) -> 1;
		(N) when N > 0 -> N * fact(N-1)
	end.
	`, env, pid)
	if err != nil {
		panic(err)
	}

	for _, n := range []int{10, 100, 1_000} {
		code := fmt.Sprintf("fact(%d).", n)
		b.Run(code, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := core.ParseEval(code, env, pid)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

func BenchmarkFibo(b *testing.B) {
	env := core.NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	_, err := core.ParseEval(`
	fun fibo
		(0) -> 0;
		(1) -> 1;
		(N) when N > 0 ->
			fibo(N-1) + fibo(N-2)
	end.
	`, env, pid)
	if err != nil {
		panic(err)
	}

	for _, n := range []int{5, 10, 15, 20} {
		code := fmt.Sprintf("fibo(%d).", n)
		b.Run(code, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := core.ParseEval(code, env, pid)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}
