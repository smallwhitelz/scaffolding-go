package orm

import (
	"context"
	"scaffolding-go/orm/internal/errs"
	"scaffolding-go/orm/model"

	"strings"
)

type Selector[T any] struct {
	table string
	model *model.Model
	where []Predicate
	sb    *strings.Builder
	args  []any

	// 可以放r、也可以直接放db
	db *DB
	//r *registry
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		sb: &strings.Builder{},
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.model, err = s.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	sb := s.sb
	sb.WriteString("SELECT * FROM ")
	// 我怎么把表名拿到
	if s.table == "" {
		sb.WriteByte('`')
		sb.WriteString(s.model.TableName)
		sb.WriteByte('`')
	} else {
		//segs := strings.Split(s.table, ".")
		//sb.WriteByte('`')
		//sb.WriteString(segs[0])
		//sb.WriteByte('`')
		//sb.WriteByte('.')
		//sb.WriteByte('`')
		//sb.WriteString(segs[1])
		//sb.WriteByte('`')
		sb.WriteString(s.table)
	}
	if len(s.where) > 0 {
		sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err := s.buildExpression(p); err != nil {
			return nil, err
		}
	}
	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case nil:
	case Predicate:
		// 在这里处理 p
		// p.left 构建好
		// p.op 构建好
		// p.right 构建好
		_, ok := exp.left.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')
		_, ok = exp.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

	case Column:
		fd, ok := s.model.FieldMap[exp.name]
		// 字段不对，或者说列不对
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.ColName)
		s.sb.WriteByte('`')
		// 剩下不考虑
	case value:
		s.sb.WriteByte('?')
		s.AddArg(exp.val)
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
	return nil
}

func (s *Selector[T]) AddArg(val any) {
	if s.args == nil {
		s.args = make([]any, 0, 4)
	}
	s.args = append(s.args, val)
}

func (s *Selector[T]) FROM(table string) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

//func (s *Selector[T]) GetV1(ctx context.Context) (*T, error) {
//	q, err := s.Build()
//	// 这个是构造sql失败
//	if err != nil {
//		return nil, err
//	}
//
//	db := s.db.db
//	// 在这里发起查询，并且处理结果集
//	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
//	// 这个是查询的错误
//	if err != nil {
//		return nil, err
//	}
//	// 你要确认有没有数据
//	if !rows.Next() {
//		// 要不要返回error?
//		// 返回error 和 sql包语义保持一致
//		return nil, ErrNoRows
//	}
//	// 在这里继续处理结果集
//	// 我怎么知道你 SELECT 出来了那些列
//	// 拿到了 SELECT 出来的列
//	cs, err := rows.Columns()
//	if err != nil {
//		return nil, err
//	}
//
//	var vals []any
//	tp := new(T)
//	// 起始地址
//	address := reflect.ValueOf(tp).UnsafePointer()
//	for _, c := range cs {
//		// c 是列名
//		fd, ok := s.model.ColumnMap[c]
//		if !ok {
//			return nil, errs.NewErrUnknownColumn(c)
//		}
//		// 是不是要计算字段的地址?
//		// 起始地址 + 偏移量
//		fdAddress := unsafe.Pointer(uintptr(address) + fd.Offset)
//		// 反射在特定的地址上，创建一个特定类型的实例
//		// 这里创建的实例是原本类型的指针类型
//		// 例如 fd.Type = int 那么val 就是 *int
//		val := reflect.NewAt(fd.Typ, fdAddress)
//		vals = append(vals, val.Interface())
//	}
//	err = rows.Scan(vals...)
//	return tp, err
//}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	// 这个是构造sql失败
	if err != nil {
		return nil, err
	}

	db := s.db.db
	// 在这里发起查询，并且处理结果集
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	// 这个是查询的错误
	if err != nil {
		return nil, err
	}
	// 你要确认有没有数据
	if !rows.Next() {
		// 要不要返回error?
		// 返回error 和 sql包语义保持一致
		return nil, ErrNoRows
	}

	tp := new(T)
	val := s.db.creator(s.model, tp)
	err = val.SetColumns(rows)
	// 接口定义好后就两件事，一个是用新接口的方法改造上层，
	// 一个就是提供不同的实现
	return tp, err

}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db
	// 在这里发起查询，并且处理结果集
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	// 在这里继续处理结果集
	for rows.Next() {
		// 构造 []*T
	}
	panic("im")
}
