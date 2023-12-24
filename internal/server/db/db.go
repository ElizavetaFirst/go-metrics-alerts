package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TestMetric struct {
	Name  string
	Type  string
	Value float64
}

type DB struct {
	conn *pgxpool.Pool
}

func (db *DB) Ping() error {
	if err := db.conn.Ping(context.Background()); err != nil {
		return fmt.Errorf("db ping error %w", err)
	}
	return nil
}

func (db *DB) Close() error {
	return nil // pgxpool.Pool не нужно закрывать
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (int64, error) {
	tag, err := db.conn.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("ExecContext return error %w", err)
	}
	return tag.RowsAffected(), nil
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.conn.QueryRow(ctx, query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	rows, err := db.conn.Query(ctx, query, args...)
	if err != nil {
		return rows, fmt.Errorf("QueryContext return error %w", err)
	}
	return rows, nil
}

func (db *DB) CreateTable(ctx context.Context) error {
	query := `
  CREATE TABLE IF NOT EXISTS metrics (
   name text NOT NULL,
   type text NOT NULL,
   value double precision,
   delta bigint,
   PRIMARY KEY (name, type)
  );
 `

	_, err := db.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to create table %w", err)
	}

	return nil
}

func NewDB(ctx context.Context, dataSourceName string) (*DB, error) {
	conn, err := pgxpool.Connect(ctx, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("can't open database %w", err)
	}
	return &DB{conn: conn}, nil
}
