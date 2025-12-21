package orm

import "scaffolding-go/orm/internal/errs"

// ErrNoRows 通过别名形式将内部错误，暴露在外面
var ErrNoRows = errs.ErrNoRows
