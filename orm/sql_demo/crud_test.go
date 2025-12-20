package sql_demo

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()
	// 这里就可以用db
	//sql.OpenDB()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// 除了 SELECT 语句，都是使用 ExecContext
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)

	// 完成了建表
	require.NoError(t, err)

	// 使用 ？ 作为查询的参数的占位符
	res, err := db.ExecContext(ctx, "INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`) VALUES(?, ?, ?, ?)",
		1, "Tom", 18, "Jerry")
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	log.Println("受影响的行数: ", affected)
	lastInsertId, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println("最后插入的ID: ", lastInsertId)
	row := db.QueryRowContext(ctx,
		"SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?", 1)
	require.NoError(t, row.Err())
	tm := TestModel{}
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	require.NoError(t, err)
	//log.Println(tm.LastName.String)
	row = db.QueryRowContext(ctx,
		"SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?", 2)
	// 这里不会有错误
	require.NoError(t, row.Err())
	tm = TestModel{}
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	// 这里报错
	require.Error(t, sql.ErrNoRows, err)
	// 认为你可能是批量查询
	rows, err := db.QueryContext(ctx,
		"SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?", 1)
	require.NoError(t, row.Err())
	for rows.Next() {
		tm = TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
		require.NoError(t, err)
		log.Println(tm)
	}

	cancel()
}

func TestTransaction(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()
	// 这里就可以用db
	//sql.OpenDB()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// 除了 SELECT 语句，都是使用 ExecContext
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)
	require.NoError(t, err)
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	require.NoError(t, err)

	// 使用 ？ 作为查询的参数的占位符
	res, err := tx.ExecContext(ctx, "INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`) VALUES(?, ?, ?, ?)",
		1, "Tom", 18, "Jerry")
	if err != nil {
		// 回滚
		err = tx.Rollback()
		if err != nil {
			log.Println(err)
		}
		return
	}
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	log.Println("受影响的行数: ", affected)
	lastInsertId, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println("最后插入的ID: ", lastInsertId)
	// 提交事务
	err = tx.Commit()
	require.NoError(t, err)

	cancel()
}

func TestPrepareStatement(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)

	// 完成了建表
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	stmt, err := db.PrepareContext(ctx, "SELECT * FROM `test_model` WHERE `id`=?")
	require.NoError(t, err)
	// id = 1
	rows, err := stmt.QueryContext(ctx, 1)
	require.NoError(t, err)
	for rows.Next() {
		tm := TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
		require.NoError(t, err)
		log.Println(tm)
	}
	cancel()
	// 整个应用关闭的时候调用
	//stmt.Close()

	// stmt, err = db.PrepareContext(ctx,
	// 	"SELECT * FROM `test_model` WHERE `id` IN (?, ?, ?)")
	// stmt, err = db.PrepareContext(ctx,
	// 	"SELECT * FROM `test_model` WHERE `id` IN (?, ?, ?, ?)")
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
