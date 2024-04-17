package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/twolodzko/goer/core/envir"
	"github.com/twolodzko/goer/core/errors"
	"github.com/twolodzko/goer/core/pids"
	"github.com/twolodzko/goer/parser"
	"github.com/twolodzko/goer/types"
	. "github.com/twolodzko/goer/types"
)

func TestEvalTerms(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    Expr
		expected Expr
	}{
		{Int(1), Int(1)},
		{Bool(true), Bool(true)},
		{Atom("foo"), Atom("foo")},
		{List{}, List{}},
		{List{[]Expr{List{}}}, List{[]Expr{List{}}}},
		{List{[]Expr{Tuple{}}}, List{[]Expr{Tuple{}}}},
		{List{[]Expr{Int(1), Bool(true), Atom("foo")}}, List{[]Expr{Int(1), Bool(true), Atom("foo")}}},
		{Tuple{}, Tuple{}},
		{Tuple{[]Expr{Int(1), Bool(true), Atom("foo")}}, Tuple{[]Expr{Int(1), Bool(true), Atom("foo")}}},
		{Tuple{[]Expr{Tuple{}}}, Tuple{[]Expr{Tuple{}}}},
		{Tuple{[]Expr{List{}}}, Tuple{[]Expr{List{}}}},
	}

	for _, tt := range testCases {
		func() {
			env := envir.EmptyEnv()
			pid := pids.NewPid()
			defer pid.Close()

			result, err := Eval(tt.input, env, pid)
			if err != nil {
				t.Errorf("evaluating '%v' resulted in an unexpected error: %s", tt.input, err)
			} else if !cmp.Equal(result, tt.expected) {
				t.Errorf("evaluating '%v' should return %v, we got %v", tt.input, tt.expected, result)
			}
		}()
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		lhs, rhs Expr
		isMatch  bool
	}{
		{Dummy{}, Dummy{}, true},
		{Dummy{}, Int(42), true},
		{Variable("X"), Dummy{}, true},
		{Dummy{}, Variable("X"), true},
		{Bool(false), Dummy{}, true},
		{Int(1), Int(1), true},
		{Int(1), Int(2), false},
		{Int(1), Bool(true), false},
		{Bool(false), Bool(false), true},
		{Bool(true), Bool(false), false},
		{Bool(false), Int(0), false},
		{Tuple{}, Tuple{}, true},
		{Tuple{}, List{}, false},
		{List{}, List{}, true},
		{Tuple{[]Expr{Int(1), Int(2), Int(3)}}, Tuple{[]Expr{Int(1), Int(2), Int(3)}}, true},
		{Tuple{[]Expr{Int(1), Int(2), Int(3)}}, Tuple{[]Expr{Int(1), Int(2)}}, false},
		{Tuple{[]Expr{Int(1), Int(2), Int(3)}}, Tuple{[]Expr{Int(2), Int(3)}}, false},
		{Tuple{[]Expr{Int(1), Int(2), Int(3)}}, Tuple{[]Expr{Int(1), Int(3), Int(2)}}, false},
		{Variable("X"), Int(1), true},
		{Int(1), Variable("X"), true},
	}
	for _, tt := range testCases {
		func() {
			env := NewEnv()
			pid := pids.NewPid()
			defer pid.Close()

			err := match(tt.lhs, tt.rhs, env, pid)
			if (err == nil) != tt.isMatch {
				t.Errorf("match %v = %v resulted in an unexpected error: %s", tt.lhs, tt.rhs, err)
			}
		}()
	}
}

func TestParseEval(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected Expr
	}{
		{"1.", Int(1)},
		{"true.", Bool(true)},
		{"foo.", Atom("foo")},
		{"{}.", Tuple{}},
		{"[].", List{}},
		{"-6.", Int(-6)},
		{"+17.", Int(17)},
		{"{1,-2, not true}.", Tuple{[]Expr{Int(1), Int(-2), Bool(false)}}},
		{"2+3.", Int(5)},
		{"2-3.", Int(-1)},
		{"4/2.", Int(2)},
		{"1==1.", Bool(true)},
		{"1==2.", Bool(false)},
		{"[1,{2,4-1}] == [1,{1+1,3}].", Bool(true)},
		{"[1,2,3] != [1,2,3,4].", Bool(true)},
		{"[1,{2,3}] != [1,{2,3}].", Bool(false)},
		{"foo == bar.", Bool(false)},
		{"foo == 1.", Bool(false)},
		{"[] == [].", Bool(true)},
		{"[2/2,2,3] == [1,1+1,6/2].", Bool(true)},
		{"{} == {}.", Bool(true)},
		{"1 < 1+1.", Bool(true)},
		{"1 > 1.", Bool(false)},
		{"2+1 <= 6/2.", Bool(true)},
		{"6/3 >= 4/2/1.", Bool(true)},
		{"16 rem 5.", Int(1)},
		{"(20 + 3) rem (12 / 2).", Int(5)},
		{"true and true.", Bool(true)},
		{"true and false.", Bool(false)},
		{"false and true.", Bool(false)},
		{"false and false.", Bool(false)},
		{"true or true.", Bool(true)},
		{"true or false.", Bool(true)},
		{"false or true.", Bool(true)},
		{"false or false.", Bool(false)},
		{"1==0 or 1+1==2.", Bool(true)},
		{"1==1 and 1==2.", Bool(false)},
		{"_ = _.", Bool(true)},
		{"X = 1.", Bool(true)},
		{"foo = X.", Bool(true)},
		{"X = _.", Bool(true)},
		{"1 = 1.", Bool(true)},
		{"2 = (((1+1))).", Bool(true)},
		{"2+2 = 4.", Bool(true)},
		{"4 = 2+2.", Bool(true)},
		{"if true -> 1+2 end = 3.", Bool(true)},
		{"6/2 = if true -> 1+2 end.", Bool(true)},
		{"{[2+2], X, {[foo,4,_]}} = {[4], (7-3), {[foo,X,false]}}.", Bool(true)},
		{"(4 + 2) / 3.", Int(2)},
		{"(foo).", Atom("foo")},
		{"{1, X, [3], _, []} = {1, 2, [Y], {4, 5}, _}.", Bool(true)},
		{"if true -> 1 end.", Int(1)},
		{"if false -> wrong; _ -> ok end.", Atom("ok")},
		{"if _ -> ok; _ -> wrong end.", Atom("ok")},
		{"if 2+2 == 4 -> ok end.", Atom("ok")},
		{"case 1 of 1 -> ok end.", Atom("ok")},
		{"case 5 of X when X > 0, X < 3 -> wrong; X when X > 3 -> ok end.", Atom("ok")},
		{"case {1, 2} of {1, 3} -> wrong; {_, 2} -> ok end.", Atom("ok")},
		{"try 1/0 recover nan end.", Atom("nan")},
		{"try 10/2 recover nan end.", Int(5)},
		{"(fun() -> ok end)().", Atom("ok")},
		{"(fun(X) -> X+1 end)(1).", Int(2)},
		{"(fun(X) -> Y=X+1, 2*X+Y end)(2).", Int(7)},
		{"(fun (X) when X < 0 -> negative; (X) when X >= 0 -> positive end)(-5).", Atom("negative")},
		{"(fun (X) when X < 0 -> negative; (X) when X >= 0 -> positive end)(15).", Atom("positive")},
		{"len([]).", Int(0)},
		{"len([1,1+2,[]]).", Int(3)},
		{"nth([1], 0).", Int(1)},
		{"nth([1,2,3], 2).", Int(3)},
		{"[] ++ [].", List{}},
		{"[1,2] ++ [3].", List{[]Expr{Int(1), Int(2), Int(3)}}},
		{"[] ++ [1].", List{[]Expr{Int(1)}}},
		{"[] ++ 1.", List{[]Expr{Int(1)}}},
		{"[1] ++ 2.", List{[]Expr{Int(1), Int(2)}}},
		{"[] ++ 1 ++ 2.", List{[]Expr{Int(1), Int(2)}}},
		{`"" ++ "".`, String("")},
		{`"\"Hello" ++ ", " ++ "World!\"".`, String(`"Hello, World!"`)},
		{"last([1]).", Int(1)},
		{"last([1,2,3]).", Int(3)},
		{"rest([1]).", List{}},
		{"rest([1,2,3]).", List{[]Expr{Int(1), Int(2)}}},
		{"rev([]).", List{}},
		{"rev([1,2,3]).", List{[]Expr{Int(3), Int(2), Int(1)}}},
		{"is_atom(foo).", Bool(true)},
		{"is_atom(true).", Bool(false)},
		{"is_int(42).", Bool(true)},
		{"is_int(foo).", Bool(false)},
		{"is_list([]).", Bool(true)},
		{"is_list([1,2,3]).", Bool(true)},
		{"is_list({[]}).", Bool(false)},
		{"is_tuple({}).", Bool(true)},
		{"is_tuple({1,foo,2+2}).", Bool(true)},
		{"is_tuple([{}]).", Bool(false)},
		{`is_str("").`, Bool(true)},
		{`is_str("yes!").`, Bool(true)},
		{"is_str(string).", Bool(false)},
		{`split("").`, List{}},
		{`split("abc").`, List{[]Expr{String("a"), String("b"), String("c")}}},
		{`str("hello").`, String("hello")},
		{"str(42).", String("42")},
		{`str("foo").`, String("foo")},
		{"str(2 + 3).", String("5")},
		{"str([1,1+1,1+2]).", String("[1,2,3]")},
		{`str({1,[2],"3"}).`, String(`{1,[2],"3"}`)},
		{`print(foo).`, String("foo")},
		{`print("Hello, World!").`, String("Hello, World!")},
		{`print({1,[],"x",true}).`, String(`{1,[],"x",true}`)},
		{`include("../examples/hello.ge").`, String("Hello, World!")},
	}

	for _, tt := range testCases {
		func() {
			env := NewEnv()
			pid := pids.NewPid()
			defer pid.Close()

			result, err := ParseEval(tt.input, env, pid)
			if err != nil {
				t.Errorf("evaluating '%s' resulted in an error: %s", tt.input, err)
			} else if !cmp.Equal(result, tt.expected) {
				t.Errorf("evaluating '%s' returned %v while we expected %v", tt.input, result, tt.expected)
			}
		}()
	}
}

func TestParseEvalErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		err   error
	}{
		{"_.", errors.Unbound{"_"}},
		{"1 + X.", errors.Unbound{"X"}},
		{"1 == _.", errors.Unbound{"_"}},
		{"-foo.", errors.NotNumber{Atom("foo")}},
		{"not 1.", errors.NotBoolean{Int(1)}},
		{"2 + x.", errors.NotNumber{Atom("x")}},
		{"1 and true.", errors.NotBoolean{Int(1)}},
		{"false or 2.", errors.NotBoolean{Int(2)}},
		{"1 = 2.", errors.NoMatch{Int(1), Int(2)}},
		{"1 / (1 - 1).", errors.DivisionByZero{}},
		{"17 rem (5 + 5 - 20 / 2).", errors.DivisionByZero{}},
		{"-(1/0).", errors.DivisionByZero{}},
		{"(1/0) + 5.", errors.DivisionByZero{}},
		{"print(str(1/0)).", errors.DivisionByZero{}},
		{"X = X.", errors.Unbound{"X"}},
		{"if false -> false end.", errors.NoTrueBranch{}},
		{"if foo -> bar end.", errors.NotBoolean{Atom("foo")}},
		{"(fun(X) -> X end)().", errors.NoFunBranch{}},
		{"(fun() -> nothing end)(1,2,3).", errors.NoFunBranch{}},
		{"len([1], [2,3]).", errors.WrongNumberArgs{}},
		{"rev(foo).", errors.NotList{Atom("foo")}},
		{"last(foo).", errors.NotList{Atom("foo")}},
		{"last([]).", errors.EmptyList{}},
		{"rest([]).", errors.EmptyList{}},
		{"rest(foo).", errors.NotList{Atom("foo")}},
		{`"hi" ++ 42.`, errors.NotString{Int(42)}},
		{"split(foo).", errors.NotString{Atom("foo")}},
		{"foo ! {1,2}.", errors.Custom{"foo is not a pid"}},
		{"case 5 of X when is_atom(X) -> atom; X when is_str(X) -> string end.", errors.NoTrueBranch{}},
		{"spawn(foo).", errors.NotFunction{Atom("foo")}},
		{"receive after xxx -> wrong end.", errors.NotNumber{Atom("xxx")}},
		{"fun f()->1 end, fun f()->2 end.", errors.Custom{"f already exists"}},
		{"(true)(5, 7).", errors.NotFunction{Bracket{Bool(true)}}},
		//                 +--------(4=X)----------+
		//        +--------|---(X=7)------+        |
		//        v        v              v        v
		{"{[2+2], X, {[foo,4,_]}} = {[4], 7, {[foo,X,false]}}.", errors.NoMatch{Variable("X"), Int(4)}},
		{"nth(wrong, 1).", errors.NotList{Atom("wrong")}},
		{"nth([], wrong).", errors.NotNumber{Atom("wrong")}},
		{"nth([], -1).", errors.Custom{"invalid index"}},
		{"nth([], 0).", errors.Custom{"invalid index"}},
		{"nth([], 1).", errors.Custom{"invalid index"}},
		{"nth([1,2,3], -1).", errors.Custom{"invalid index"}},
		{"nth([1,2,3], 3).", errors.Custom{"invalid index"}},
		{"error(wrong).", errors.NotString{Atom("wrong")}},
		{`error("hello!").`, errors.Custom{"hello!"}},
	}

	for _, tt := range testCases {
		func() {
			env := NewEnv()
			pid := pids.NewPid()
			defer pid.Close()

			_, err := ParseEval(tt.input, env, pid)
			if !cmp.Equal(err, tt.err) {
				t.Errorf("evaluating '%s' should throw error: %s, but it thrown: %s", tt.input, tt.err, err)
			}
		}()
	}
}

func TestPartialEval(t *testing.T) {
	t.Parallel()

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	exprs, err := parser.Parse("1+1, Y=2+2, 3+3.")
	if err != nil {
		t.Errorf("unexpected parsing error: %s", err)
	}

	// return last expression without evaluating
	last, _, err := partialEval(exprs, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(last, exprs[2]) {
		t.Errorf("expected %v, got %v", exprs[2], last)
	}

	// if everything before was evaluated, this should exist
	expected := Int(4)
	result, err := Eval(Variable("Y"), env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestMatchSet(t *testing.T) {
	t.Parallel()
	var err error

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	// fresh values
	_, err = ParseEval("X = 1.", env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	_, err = ParseEval("Y = X.", env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// they are equal
	_, err = ParseEval("X = Y.", env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// this match should fail
	_, err = ParseEval("X = 2.", env, pid)
	expectedErr := errors.NoMatch{Variable("X"), Int(2)}
	if !cmp.Equal(err, expectedErr) {
		t.Errorf("expected error: %s, got %s", expectedErr, err)
	}
}

func TestNamedFun(t *testing.T) {
	t.Parallel()

	expected := Atom("ok")
	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	result, err := ParseEval(`
	fun identity(X) -> X end,
	identity(ok).
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expected) {
		t.Errorf("expected message %v, got %v", expected, result)
	}
}

func TestTailCallOptimization(t *testing.T) {
	t.Parallel()

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	// this code stack-overflows without tail-call optimization
	largeNumberOfIterations := 1_000_000
	code := fmt.Sprintf(
		`fun down
			(0) -> ok;
			(X) -> down(X - 1)
		 end,
		 down(%d).`,
		largeNumberOfIterations,
	)
	_, err := ParseEval(code, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestIterating(t *testing.T) {
	t.Parallel()

	var (
		err              error
		result, expected types.Expr
	)
	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	_, err = ParseEval(`
	fun reverse
		%% interface
		(Lst) -> reverse(Lst, []);
		%% implementation
		([], Acc) -> Acc;
		(Lst, Acc) -> reverse(rest(Lst), Acc ++ [last(Lst)])
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	expected = types.List{[]types.Expr{types.Int(3), types.Int(2), types.Int(1)}}
	result, err = ParseEval("reverse([1,2,3]).", env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSleep(t *testing.T) {
	t.Parallel()

	waitTime := 1500
	expectedResult := Int(waitTime)

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	startTime := time.Now()
	result, err := ParseEval(fmt.Sprintf("sleep(%d).", waitTime), env, pid)
	duration := time.Since(startTime)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expectedResult) {
		t.Errorf("expected %v, got %v", expectedResult, result)
	}

	// there may be an overhead, but it should not be lower than the expected wait time
	expectedDuration := time.Duration(waitTime * int(time.Millisecond))
	if duration < expectedDuration {
		t.Errorf("the duration was %d msec < %d msec expected", duration.Milliseconds(), expectedDuration.Milliseconds())
	}
}

func TestExit(t *testing.T) {
	t.Parallel()

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	_, err := ParseEval("exit(reason).", env, pid)
	expectedErr := errors.Exit{Atom("reason")}
	if !cmp.Equal(err, expectedErr) {
		t.Errorf("expected error: '%s', got '%s'", expectedErr, err)
	}
}

func TestSelf(t *testing.T) {
	t.Parallel()

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	result, err := ParseEval("self().", env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	gotPid, ok := result.(pids.Pid)
	if !ok {
		t.Errorf("Expected pid, got %T", result)
	} else if !cmp.Equal(pid, gotPid) {
		t.Errorf("expected pid %v, got %v", pid, gotPid)
	}
}

func TestSendMessage(t *testing.T) {
	t.Parallel()

	expected := Atom("hi")
	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	// send a message to yourself
	_, err := ParseEval("self() ! hi.", env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// the message is received
	msg := <-pid.Messages()
	if !cmp.Equal(msg, expected) {
		t.Errorf("expected %v, got %v", expected, msg)
	}
}

func TestMessageFromSelf(t *testing.T) {
	t.Parallel()

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	expected := Tuple{[]Expr{pid, Atom("hello")}}
	result, err := ParseEval(`
	self() ! hello,
	receive
		Msg -> {self(), Msg}
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expected) {
		t.Errorf("expected message %v, got %v", expected, result)
	}
}

func TestCommunicate(t *testing.T) {
	t.Parallel()

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	expected := Tuple{[]Expr{Atom("ack"), Atom("hi")}}

	result, err := ParseEval(`
	Pid = spawn(fun() ->
		receive
			{Sender, Msg} ->
				Sender ! {ack, Msg}
		end
	end),

	Pid ! {self(), hi},

	receive
		Msg -> Msg
	after
		100 -> timeout
	end.
	`, env, pid)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expected) {
		t.Errorf("expected message %v, got %v", expected, result)
	}
}

func TestReceiveTimeout(t *testing.T) {
	t.Parallel()

	waitTime := 1500
	expectedResult := Atom("ok")
	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	startTime := time.Now()
	result, err := ParseEval(fmt.Sprintf("receive after %d -> ok end.", waitTime), env, pid)
	duration := time.Since(startTime)

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expectedResult) {
		t.Errorf("expected %v, got %v", expectedResult, result)
	}

	// there may be an overhead, but it should not be lower than the expected wait time
	expectedDuration := time.Duration(waitTime * int(time.Millisecond))
	if duration < expectedDuration {
		t.Errorf("the duration was %d msec < %d msec expected", duration.Milliseconds(), expectedDuration.Milliseconds())
	}
}

func TestTimeout(t *testing.T) {
	t.Parallel()

	env := NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	startTime := time.Now()
	result, err := ParseEval(`
	Root = self(),

	fun loop() ->
		receive
			keep_awake ->
				Root ! ok,
				loop()
		after
			200 ->
				Root ! timeout
		end
	end,

	Pid = spawn(fun() -> loop() end),

	sleep(100),
	Pid ! keep_awake,
	receive
		ok -> ok
	end,

	sleep(100),
	Pid ! keep_awake,
	receive
		ok -> ok
	end,

	receive
		Msg -> Msg
	end.
	`, env, pid)
	duration := time.Since(startTime)

	expectedResult := Atom("timeout")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !cmp.Equal(result, expectedResult) {
		t.Errorf("expected message %v, got %v", expectedResult, result)
	}

	// there may be an overhead, but it should not be lower than the expected wait time
	expectedDuration := time.Duration(400 * time.Millisecond)
	if duration < expectedDuration {
		t.Errorf("the duration was %d msec < %d msec expected", duration.Milliseconds(), expectedDuration.Milliseconds())
	}
}
