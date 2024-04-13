package types

type (
	Expr     = any
	Bool     bool
	Int      int
	String   string
	Atom     string
	Variable string
)

// Dummy variable ("_" in Erlang). In pattern matching it matches anything.
type Dummy struct{}

type Tuple struct {
	Values []Expr
}

type List struct {
	Values []Expr
}

func (l List) Len() int {
	return len(l.Values)
}

func (l List) Append(exprs ...Expr) List {
	return List{append(l.Values, exprs...)}
}

// Expression enclosed in brackets.
type Bracket struct {
	Expr Expr
}

// Unary operation.
type UnaryOperation struct {
	Op  string
	Rhs Expr
}

// Binary operation.
type BinaryOperation struct {
	Op       string
	Lhs, Rhs Expr
}

// Call an expression (anonymous function, name of the function)
// with the arguments.
type Call struct {
	Callable Expr
	Args     []Expr
}

// A function definition, consisting of one or more branches executed conditionally.
// The branches are picked by pattern matching their arguments.
type Definition struct {
	Name     string
	Branches []FunBranch
}

// A representation of function branch (part of the function definition).
type FunBranch struct {
	Args   []Expr
	Guards []Expr
	Body   []Expr
}

// If statement.
type If struct {
	Branches []IfBranch
}

// Body of expressions to be executed conditionally.
type IfBranch struct {
	Cond Expr
	Body []Expr
}

// Case statement.
type Case struct {
	Arg      Expr
	Branches []PatternBranch
}

// Body of expressions to be executed conditionally.
type PatternBranch struct {
	Pattern Expr
	Guards  []Expr
	Body    []Expr
}

// Receive statement.
type Receive struct {
	Branches []PatternBranch
	After    IfBranch
}

// TryRecover-recover statement.
type TryRecover struct {
	Body    []Expr
	Recover []Expr
}
