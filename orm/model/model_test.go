package model

import (
	"database/sql"
	"reflect"
	"scaffolding-go/orm/internal/errs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_registry_Register(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		wantErr   error
	}{
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},

		{
			name:    "map",
			entity:  map[string]string{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			entity:  []int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "basic type",
			entity:  123,
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				Fields: []*Field{
					{
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					{
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					{
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
			},
		},
	}
	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, f := range tc.wantModel.Fields {
				fieldMap[f.GoName] = f
				columnMap[f.ColName] = f
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func TestRegister_get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		wantErr   error
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				Fields: []*Field{
					{
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					{
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					{
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
			},
		},
		{
			name: "tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column=first_name_t"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
				Fields: []*Field{
					{
						ColName: "first_name_t",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},

		{
			name: "empty column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column="`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name: "column only",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},

		{
			name: "ignore atg",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"abc=abc"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name:   "table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name_t",
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},

		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr_t",
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},

		{
			name:   "table name empty",
			entity: &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
	}
	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, f := range tc.wantModel.Fields {
				fieldMap[f.GoName] = f
				columnMap[f.ColName] = f
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
			assert.Equal(t, tc.wantModel, m)
			typ := reflect.TypeOf(tc.entity)
			cache, ok := r.(*registry).models.Load(typ)
			assert.True(t, ok)
			assert.Equal(t, tc.wantModel, cache)
		})
	}
}

type CustomTableName struct {
	FirstName string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	FirstName string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}

type EmptyTableName struct {
	FirstName string
}

func (c *EmptyTableName) TableName() string {
	return ""
}

func Test_underscoreName(t *testing.T) {
	testCases := []struct {
		name    string
		srcStr  string
		wantStr string
	}{
		// 我们这些用例就是为了确保
		// 在忘记 underscoreName 的行为特性之后
		// 可以从这里找回来
		// 比如说过了一段时间之后
		// 忘记了 ID 不能转化为 id
		// 那么这个测试能帮我们确定 ID 只能转化为 i_d
		{
			name:    "upper cases",
			srcStr:  "ID",
			wantStr: "i_d",
		},
		{
			name:    "use number",
			srcStr:  "Table1Name",
			wantStr: "table1_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := underscoreName(tc.srcStr)
			assert.Equal(t, tc.wantStr, res)
		})
	}
}

func TestModelWithTableName(t *testing.T) {
	r := NewRegistry()
	m, err := r.Register(&TestModel{}, WithTableName("test_model_ttt"))
	require.NoError(t, err)
	assert.Equal(t, "test_model_ttt", m.TableName)
}

func TestModelWithColumnName(t *testing.T) {
	testCases := []struct {
		name    string
		field   string
		colName string

		wantColName string
		wantErr     error
	}{
		{
			name:        "column name",
			field:       "FirstName",
			colName:     "first_name_ccc",
			wantColName: "first_name_ccc",
		},
		{
			name:    "unknow column",
			field:   "Address",
			colName: "address_ccc",
			wantErr: errs.NewErrUnknownField("Address"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			m, err := r.Register(&TestModel{}, WithColumnName(tc.field, tc.colName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd, ok := m.FieldMap[tc.field]
			require.True(t, ok)
			assert.Equal(t, tc.wantColName, fd.ColName)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
