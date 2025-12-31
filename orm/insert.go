package orm

import (
	"reflect"
	"scaffolding-go/orm/internal/errs"
	"scaffolding-go/orm/model"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	builder
	values  []*T
	db      *DB
	columns []string
	//onDuplicateKey []Assignable
	onDuplicateKey *Upsert
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
		db: db,
	}
}

//func (i *Inserter[T]) Upsert(assigns ...Assignable) *Inserter[T] {
//	i.onDuplicateKey = assigns
//	return i
//}

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

// Values 指定插入的数据
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	i.sb.WriteString("INSERT INTO ")
	m, err := i.db.r.Get(i.values[0])
	i.model = m
	if err != nil {
		return nil, err
	}
	// 拼接表名
	i.quote(m.TableName)
	// 一定要显示的指定列的顺序，不然我们不知道数据库中默认的顺序
	// 我们要构造列的名字，类似 `test_model`(col1,col2)
	i.sb.WriteByte('(')
	fields := m.Fields
	// 用户指定了列
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := m.FieldMap[fd]
			// 传入了乱七八糟的列
			if !ok {
				return nil, errs.NewErrUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}
	// 不能遍历这个map，因为在go里面，map的遍历每一次都不一样
	// 所以要额外的引入一个 Fields
	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(field.ColName)
	}
	i.sb.WriteByte(')')
	// 拼接 Values
	i.sb.WriteString(" VALUES ")

	// 预估参数数量是：我有多少行乘以我有多少个字段
	i.args = make([]any, 0, len(i.values)*len(fields))
	for j, val := range i.values {
		if j > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteString("?")
			// 把参数读出来
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			i.addArg(arg)
		}
		i.sb.WriteByte(')')
	}
	if i.onDuplicateKey != nil {
		err := i.dialect.buildUpsert(&i.builder, i.onDuplicateKey)
		if err != nil {
			return nil, err
		}
	}
	i.sb.WriteByte(';')
	return &Query{SQL: i.sb.String(), Args: i.args}, nil
}
