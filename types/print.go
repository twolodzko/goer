package types

import (
	"fmt"
	"strings"
)

func (s String) String() string {
	return fmt.Sprintf("\"%s\"", string(s))
}

func (d Dummy) String() string {
	return "_"
}

func (t Tuple) String() string {
	return fmt.Sprintf("{%s}", stringify(t.Values))
}

func (l List) String() string {
	return fmt.Sprintf("[%s]", stringify(l.Values))
}

func (b Bracket) String() string {
	return fmt.Sprintf("(%v)", b.Expr)
}

func (o UnaryOperation) String() string {
	return fmt.Sprintf("%s %v", o.Op, o.Rhs)
}

func (o BinaryOperation) String() string {
	return fmt.Sprintf("%v %s %v", o.Lhs, o.Op, o.Rhs)
}

func (c Call) String() string {
	return fmt.Sprintf("%v(%v)", c.Callable, stringify(c.Args))
}

func (d Definition) String() string {
	var s []string
	for _, branch := range d.Branches {
		s = append(s, fmt.Sprint(branch))
	}
	body := strings.Join(s, "; ")
	if d.Name != "" {
		return fmt.Sprintf("fun %s %s end", d.Name, body)
	}
	return fmt.Sprintf("fun %s end", body)
}

func (b FunBranch) String() string {
	if b.Guards == nil {
		return fmt.Sprintf("(%s) -> %s", stringify(b.Args), stringify(b.Body))
	}
	return fmt.Sprintf("(%s) when %s -> %s", stringify(b.Args), stringify(b.Guards), stringify(b.Body))
}

func (i If) String() string {
	var s []string
	for _, branch := range i.Branches {
		s = append(s, fmt.Sprint(branch))
	}
	return fmt.Sprintf("if %s end", strings.Join(s, "; "))
}

func (b IfBranch) String() string {
	return fmt.Sprintf("%v -> %s", b.Cond, stringify(b.Body))
}

func (c Case) String() string {
	var s []string
	for _, branch := range c.Branches {
		s = append(s, fmt.Sprint(branch))
	}
	return fmt.Sprintf("case %v of %s end", c.Arg, strings.Join(s, "; "))
}

func (b PatternBranch) String() string {
	if b.Guards == nil {
		return fmt.Sprintf("%v -> %s", b.Pattern, stringify(b.Body))
	}
	return fmt.Sprintf("%v when %s -> %s", b.Pattern, stringify(b.Guards), stringify(b.Body))
}

// Convert list of expressions to a comma-separated string representation.
func stringify(exprs []Expr) string {
	var s []string
	for _, expr := range exprs {
		s = append(s, fmt.Sprint(expr))
	}
	return strings.Join(s, ",")
}
