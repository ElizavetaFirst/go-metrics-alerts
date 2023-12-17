package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
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
	return db.conn.Ping(context.Background())
}

func (db *DB) Close() error {
	return nil // pgxpool.Pool не нужно закрывать
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (int64, error) {
	tag, err := db.conn.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.conn.QueryRow(ctx, query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.conn.Query(ctx, query, args...)
}

func (db *DB) CreateTable(ctx context.Context) error {
	fmt.Println(db.conn.Config().ConnConfig.User)
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
		return errors.Wrap(err, "unable to create table")
	}

	//grantQuery := `GRANT ALL PRIVILEGES ON TABLE metrics TO postgres;`

	//_, err = db.conn.Exec(ctx, grantQuery)
	//if err != nil {
	//	return errors.Wrap(err, "unable to grant privileges")
	//}

	return nil
}

func NewDB(ctx context.Context, dataSourceName string) (*DB, error) {
	fmt.Println(dataSourceName)
	conn, err := pgxpool.Connect(ctx, dataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "can't open database")
	}
	return &DB{conn: conn}, nil
}
