package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type FakeDB struct{}

func (f *FakeDB) Ping() error {
	return nil
}

func (f *FakeDB) Close() error {
	return nil
}

type SQLDB interface {
	Ping() error
	Close() error
}

type DB struct {
	SQLDB
}

// NewDB инициализирует и возвращает новый *DB.
func NewDB(dataSourceName string) (*DB, error) {
	fmt.Println(dataSourceName)
	realDB, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{realDB}, nil
}
