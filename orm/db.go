package orm

import "database/sql"

type DBOption func(db *DB)

// DB 是一个 sql.DB 的装饰器
type DB struct {
	r  *registry
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
		r:  newRegistry(),
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// MustOpen 创建一个DB，如果失败则会 panic
func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	res, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}
