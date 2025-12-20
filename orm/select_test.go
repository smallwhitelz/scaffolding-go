package orm

import (
	"database/sql"
	"scaffolding-go/orm/internal/errs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelector_Build(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)
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

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
