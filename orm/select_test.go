package orm

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelector_Build(t *testing.T) {
	testCases := []struct {
		name string

		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: &Selector[TestModel]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},

		{
			name:    "from",
			builder: (&Selector[TestModel]{}).FROM("test_model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},

		{
			name:    "empty from",
			builder: (&Selector[TestModel]{}).FROM(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},

		{
			name:    "with db",
			builder: (&Selector[TestModel]{}).FROM("`test_db`.`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_db`.`test_model`;",
				Args: nil,
			},
		},

		{
			name:    "empty where",
			builder: (&Selector[TestModel]{}).Where(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
		},

		{
			name:    "where",
			builder: (&Selector[TestModel]{}).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `Age` = ?;",
				Args: []any{18},
			},
		},

		{
			name:    "not",
			builder: (&Selector[TestModel]{}).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE  NOT (`Age` = ?);",
				Args: []any{18},
			},
		},

		{
			name:    "and",
			builder: (&Selector[TestModel]{}).Where(C("age").Eq(18).And(C("first_name").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},

		{
			name:    "or",
			builder: (&Selector[TestModel]{}).Where(C("age").Eq(18).Or(C("first_name").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
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
