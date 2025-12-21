package additional

import "errors"

type MyStructOption func(m *MyStruct)
type MyStructOptionErr func(m *MyStruct) error

type MyStruct struct {
	// 第一个部分是必须用户输入的字段
	id   uint64
	name string
	// 第二个部分是可选字段
	address string
	// 这里可以有很多字段
	field1 int
	field2 int
}

func WithField(field1, field2 int) MyStructOption {
	return func(m *MyStruct) {
		m.field1 = field1
		m.field2 = field2
	}
}

func WithAddress(address string) MyStructOption {
	return func(m *MyStruct) {
		m.address = address
	}
}

func WithAddressPanic(address string) MyStructOption {
	return func(m *MyStruct) {
		if address == "" {
			panic("不能为空")
		}
		m.address = address
	}
}

// NewMyStruct 参数包含所有必须用户输入的字段
func NewMyStruct(id uint64, name string, opts ...MyStructOption) *MyStruct {
	// 构造必传部分
	res := &MyStruct{
		id:   id,
		name: name,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func WithAddressErr(address string) MyStructOptionErr {
	return func(m *MyStruct) error {
		if address == "" {
			return errors.New("不能为空")
		}
		m.address = address
		return nil
	}
}

func NewMyStructV1(id uint64, name string, opts ...MyStructOptionErr) (*MyStruct, error) {
	// 构造必传部分
	res := &MyStruct{
		id:   id,
		name: name,
	}
	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
