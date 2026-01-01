package orm

import (
	"scaffolding-go/orm/internal/valuer"
	"scaffolding-go/orm/model"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       model.Registry
	mdls    []Middleware
}
