package orm

// 衍生类型
type op string

const (
	opEq  op = "="
	opLT  op = "<"
	opNot op = "NOT"
	opAnd op = "AND"
	opOr  op = "OR"
)

func (o op) String() string {
	return string(o)
}

// 这种叫做别名
//type op = string

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

// Eq("id",12)
// Eq(sub,"id","12")
// Eq(sub.id,"12")
// Eq("sub.id","12")
//func Eq(column string, right any) Predicate {
//	return Predicate{
//		Column: column,
//		Op:     "=",
//		Arg:    right,
//	}
//}

// Not(C("name").Eq("Tom"))
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// C("id").Eq(12).And(C("name").Eq("Tom"))
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left: left,
		op:   opAnd,

		right: right,
	}
}

// C("id").Eq(12).Or(C("name").Eq("Tom"))
func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

func (Predicate) expr() {

}

type value struct {
	val any
}

func (value) expr() {

}
