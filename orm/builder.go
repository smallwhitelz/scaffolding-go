package orm

import (
	"scaffolding-go/orm/internal/errs"
	"scaffolding-go/orm/model"
	"strings"
)

type builder struct {
	sb      strings.Builder
	args    []any
	model   *model.Model
	dialect Dialect
	quoter  byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) buildColumn(name string) error {
	fd, ok := b.model.FieldMap[name]
	if !ok {
		return errs.NewErrUnknownField(name)
	}
	b.quote(fd.ColName)
	return nil
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
