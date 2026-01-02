package orm

type Column struct {
	table TableReference
	name  string
	alias string
}

func (c Column) selectable() {

}

func C(name string) Column {
	return Column{
		name: name,
	}
}

func (c Column) assign() {}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
		table: c.table,
	}
}

// Eq 代表相等
// C("id").Eq(12)
// sub.C("id").Eq(12) 复杂查询，可读性高
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(arg),
	}
}

func valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{val: arg}
	}
}

func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: value{val: arg},
	}
}

func (c Column) expr() {

}
