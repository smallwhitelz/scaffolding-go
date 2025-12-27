package unsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

type UnsafeAccessor struct {
	fields  map[string]FieldMeta
	address unsafe.Pointer
}

func NewUnsafeAccessor(entity any) *UnsafeAccessor {
	typ := reflect.TypeOf(entity)
	typ = typ.Elem()
	numField := typ.NumField()
	fields := make(map[string]FieldMeta, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fields[fd.Name] = FieldMeta{
			Offset: fd.Offset,
			Typ:    fd.Type,
		}
	}
	val := reflect.ValueOf(entity)
	return &UnsafeAccessor{
		fields:  fields,
		address: val.UnsafePointer(),
	}
}

func (a *UnsafeAccessor) Field(field string) (any, error) {
	// 起始地址 + 字段偏移量
	fd, ok := a.fields[field]
	if !ok {
		return nil, errors.New("非法字段")
	}
	// 字段起始位置
	fdAdress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// 如果知道类型，就这么读
	//return *(*int)(fdAdress), nil
	// 不知道类型
	return reflect.NewAt(fd.Typ, fdAdress).Elem().Interface(), nil
}

func (a *UnsafeAccessor) SetField(field string, val any) error {
	// 起始地址 + 字段偏移量
	fd, ok := a.fields[field]
	if !ok {
		return errors.New("非法字段")
	}
	// 字段起始位置
	fdAdress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// 知道确切类型就这么写
	//*(*int)(fdAdress) = val.(int)
	// 不知道确切类型写法
	reflect.NewAt(fd.Typ, fdAdress).Elem().Set(reflect.ValueOf(val))
	return nil
}

type FieldMeta struct {
	Offset uintptr
	Typ    reflect.Type
}
