package orm

// Expression 是一个标记接口，代表表达式
type Expression interface {
	expr()
}

// RawExpr 代表的是原生表达式
// Raw 不是 Row
type RawExpr struct {
	raw  string
	args []any
}

func (r RawExpr) selectable() {}
func (r RawExpr) expr()       {}
func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}
