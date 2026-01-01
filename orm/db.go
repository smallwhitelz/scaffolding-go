package orm

import (
	"context"
	"database/sql"
	"scaffolding-go/orm/internal/valuer"
	"scaffolding-go/orm/model"
)

type DBOption func(db *DB)

// DB 是一个 sql.DB 的装饰器
type DB struct {
	core
	db *sql.DB
}

func Open(driver string, dataSourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

// OpenDB 有时候用户可能自己已经建好了db
// 方便测试
// 结合其他orm框架
func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		core: core{
			r:       model.NewRegistry(),
			creator: valuer.NewUnsafeValue,
			dialect: DialectMySQL,
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}

func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}

// MustOpen 创建一个DB，如果失败则会 panic
func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	res, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{
		tx: tx,
	}, nil
}

type txKey struct{}

// 事务的扩散方案
func (db *DB) BeginTxV2(ctx context.Context,
	opts *sql.TxOptions) (context.Context, *Tx, error) {
	val := ctx.Value(txKey{})
	tx, ok := val.(*Tx)
	// 存在一个事务并且没有被提交或者回滚
	if ok && !tx.done {
		return ctx, tx, nil
	}
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, txKey{}, tx)
	return ctx, tx, nil
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func (db *DB) getCore() core {
	return db.core
}
