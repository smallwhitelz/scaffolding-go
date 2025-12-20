package orm

type DBOption func(db *DB)

type DB struct {
	r *registry
}

func NewDB(opts ...DBOption) (*DB, error) {
	res := &DB{
		r: newRegistry(),
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// MustNewDB 创建一个DB，如果失败则会 panic
func MustNewDB(opts ...DBOption) *DB {
	res, err := NewDB(opts...)
	if err != nil {
		panic(err)
	}
	return res
}
