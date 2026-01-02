package orm

import (
	"context"
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

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	// 这个是构造sql失败
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	// 在这里发起查询，并且处理结果集
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	// 这个是查询的错误
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	// 你要确认有没有数据
	if !rows.Next() {
		// 要不要返回error?
		// 返回error 和 sql包语义保持一致
		return &QueryResult{
			Err: ErrNoRows,
		}
	}
	tp := new(T)
	val := c.creator(c.model, tp)
	err = val.SetColumns(rows)
	// 接口定义好后就两件事，一个是用新接口的方法改造上层，
	// 一个就是提供不同的实现
	return &QueryResult{
		Result: tp,
		Err:    err,
	}
}

func exec(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func execHandler(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Result: Result{
				err: err,
			},
		}
	}
	res, err := sess.execContext(ctx, q.SQL, q.Args...)
	return &QueryResult{
		Result: Result{
			err: err,
			res: res,
		},
	}
}
