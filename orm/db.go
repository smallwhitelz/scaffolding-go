package orm

import (
	"database/sql"
	"scaffolding-go/orm/internal/valuer"
	"scaffolding-go/orm/model"
)

type DBOption func(db *DB)

// DB 是一个 sql.DB 的装饰器
type DB struct {
	r       model.Registry
	db      *sql.DB
	creator valuer.Creator
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
		r:       model.NewRegistry(),
		db:      db,
		creator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
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
