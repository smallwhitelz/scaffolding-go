package sql_demo

import (
	"database/sql/driver"
	"errors"

	"github.com/bytedance/sonic"
)

type JSONColumn[T any] struct {
	Val T
	// NULL 的问题
	Valid bool
}

func (j *JSONColumn[T]) Value() (driver.Value, error) {
	// NULL 字段
	if !j.Valid {
		return nil, nil
	}
	return sonic.Marshal(j.Val)
}

func (j *JSONColumn[T]) Scan(src any) error {
	//    int64
	//    float64
	//    bool
	//    []byte
	//    string
	//    time.Time
	//    nil - for NULL values
	var bs []byte
	switch data := src.(type) {
	case string:
		bs = []byte(data)
	case []byte:
		bs = data
	case nil:
		// 说明数据库里面存的就是NULL
		return nil
	default:
		return errors.New("不支持类型")
	}
	err := sonic.Unmarshal(bs, &j.Val)
	if err == nil {
		// 代表有数据
		j.Valid = true
	}
	return err
}
