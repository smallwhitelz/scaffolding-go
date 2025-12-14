package homework_delete

import (
	"reflect"
)

type Deleter[T any] struct {
	builder
	table string
	where []Predicate
	args  []any
}

func (d *Deleter[T]) Build() (*Query, error) {
	d.sb.WriteString("DELETE FROM ")
	if d.table == "" {
		var t T
		d.sb.WriteByte('`')
		d.sb.WriteString(reflect.TypeOf(t).Name())
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteString(d.table)
	}

	// 构造 WHERE
	if len(d.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		d.sb.WriteString(" WHERE ")
		err := d.builder.buildPredicates(d.where)
		if err != nil {
			return nil, err
		}
	}
	d.sb.WriteString(";")
	return &Query{
		SQL:  d.sb.String(),
		Args: d.builder.args,
	}, nil
}

// From accepts model definition
func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}

// Where accepts predicates
func (d *Deleter[T]) Where(predicates ...Predicate) *Deleter[T] {
	d.where = predicates
	return d
}
