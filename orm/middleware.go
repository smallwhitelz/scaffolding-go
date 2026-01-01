package orm

import (
	"context"
	"scaffolding-go/orm/model"
)

type QueryContext struct {
	// 查询类型，标记crud
	Type string

	// 代表查询本身
	Builder QueryBuilder

	Model *model.Model
}

type QueryResult struct {
	// Result 在不同的查询下类型是不同的
	// Select --> []*T 或者 *T
	// 其他就是类型 Result
	Result any

	// 查询本身出的问题
	Err error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

type Middleware func(next Handler) Handler
