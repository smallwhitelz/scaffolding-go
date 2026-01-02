package orm

import (
	"scaffolding-go/orm/internal/errs"
	"strings"
)

type builder struct {
	core
	sb   strings.Builder
	args []any

	quoter byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) buildColumn(col Column) error {
	switch table := col.table.(type) {
	case nil:
		fd, ok := b.model.FieldMap[col.name]
		if !ok {
			return errs.NewErrUnknownField(col.name)
		}
		b.quote(fd.ColName)
		if col.alias != "" {
			b.sb.WriteString(" AS ")
			b.quote(col.alias)
		}
		return nil
	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}
		fd, ok := m.FieldMap[col.name]
		if !ok {
			return errs.NewErrUnknownField(col.name)
		}
		if table.alias != "" {
			b.quote(table.alias)
			b.sb.WriteByte('.')
		}
		b.quote(fd.ColName)
		if col.alias != "" {
			b.sb.WriteString(" AS ")
			b.quote(col.alias)
		}
		return nil
	default:
		return errs.NewErrUnsupportedTable(table)
	}
}

func (b *builder) addArg(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if b.args == nil {
		b.args = make([]any, 0, 4)
	}
	b.args = append(b.args, vals...)
}
