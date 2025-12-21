package orm

import (
	"context"
	"database/sql"
	"errors"
	"scaffolding-go/orm/internal/errs"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name string

		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},

		{
			name:    "from",
			builder: NewSelector[TestModel](db).FROM("test_model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},

		{
			name:    "empty from",
			builder: NewSelector[TestModel](db).FROM(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},

		{
			name:    "with db",
			builder: NewSelector[TestModel](db).FROM("`test_db`.`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_db`.`test_model`;",
				Args: nil,
			},
		},

		{
			name:    "empty where",
			builder: NewSelector[TestModel](db).Where(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},

		{
			name:    "where",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},

		{
			name:    "not",
			builder: NewSelector[TestModel](db).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
		},

		{
			name:    "and",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).And(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},

		{
			name:    "or",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},

		{
			name:    "invalid column",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("111").Eq("Tom"))),
			wantErr: errs.NewErrUnknownField("111"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

// memoryDB 返回一个基于内存的 ORM，它使用的是 sqlite3 内存模式。
func memoryDB(t *testing.T) *DB {
	orm, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	if err != nil {
		t.Fatal(err)
	}
	return orm
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	// 对应于 query error
	mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))

	// 对应于no rows
	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .* ").WillReturnRows(rows)

	// data
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// scan error
	//rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	//rows.AddRow("abc", "Tom", "18", "Jerry")
	//mock.ExpectQuery("SELECT .*").WillReturnRows(rows)
	testCases := []struct {
		name string
		s    *Selector[TestModel]

		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "invalid query",
			s:       NewSelector[TestModel](db).Where(C("xxx").Eq(1)),
			wantErr: errs.NewErrUnknownField("xxx"),
		},

		{
			name:    "query error",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: errors.New("query error"),
		},

		{
			name:    "no rows",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			s:    NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantRes: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
		//{
		//	name:    "scan error",
		//	s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
		//	wantErr: ErrNoRows,
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
