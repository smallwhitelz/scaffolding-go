package slowquery

import (
	"context"
	"errors"
	"scaffolding-go/orm"
)

type MiddlewareBuilder struct {
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			// 禁用删除操作
			if qc.Type == "DELETE" {
				return &orm.QueryResult{
					Err: errors.New("delete operations are disabled"),
				}
			}
			return next(ctx, qc)
		}
	}
}
