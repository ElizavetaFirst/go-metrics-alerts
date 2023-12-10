package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type FakeDB struct{}

type FakeResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (fr *FakeResult) LastInsertId() (int64, error) {
	return fr.lastInsertId, nil
}

func (fr *FakeResult) RowsAffected() (int64, error) {
	return fr.rowsAffected, nil
}

func (f *FakeDB) Ping() error {
	return nil
}

func (f *FakeDB) Close() error {
	return nil
}

func (f *FakeDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return &FakeResult{
		lastInsertId: 1,
		rowsAffected: 1,
	}, nil
}

func (f *FakeDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	var row *sql.Row
	return row
}

func (f *FakeDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	return rows, nil
}

type SQLDB interface {
	Ping() error
	Close() error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type DB struct {
	SQLDB
}

// NewDB инициализирует и возвращает новый *DB.
func NewDB(dataSourceName string) (*DB, error) {
	fmt.Println(dataSourceName)
	realDB, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "can't open database")
	}
	return &DB{realDB}, nil
}
