package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
)

type DBStorage struct {
	conn *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, databaseDSN string) (*DBStorage, error) {
	conn, err := newDB(ctx, databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("database is not inited: %w", ErrCantConnectDB)
	}

	dbStorage := DBStorage{
		conn: conn,
	}

	err = dbStorage.CreateTable(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't create table %w", err)
	}

	return &dbStorage, nil
}

func (dbs *DBStorage) Update(ctx context.Context, opts *UpdateOptions) error {

	var value, delta interface{}
	if opts.Update.Type == constants.Counter {
		delta = opts.Update.Value
	} else if opts.Update.Type == constants.Gauge {
		value = opts.Update.Value
	}

	_, err := dbs.conn.Exec(ctx, `
	INSERT INTO metrics (name, type, value, delta)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT(name, type) DO UPDATE
	SET value = EXCLUDED.value, 
		delta = CASE 
			WHEN metrics.type = 'counter' THEN metrics.delta + EXCLUDED.delta
			ELSE EXCLUDED.delta
		END;`,
		opts.MetricName, opts.Update.Type, value, delta)

	if err != nil {
		return fmt.Errorf("ExecContext return error %w", err)
	}
	return nil
}

func (dbs *DBStorage) Get(ctx context.Context, opts *GetOptions) (Metric, error) {
	row := dbs.conn.QueryRow(ctx, `SELECT value, delta FROM metrics WHERE name=$1 AND type=$2`,
		opts.MetricName, opts.MetricType)

	var value, delta interface{}
	err := row.Scan(&value, &delta)
	if err != nil {
		return Metric{}, fmt.Errorf("can't get metric from DBStorage %s %s: %w",
			opts.MetricName, opts.MetricType, ErrMetricNotFound)
	}

	var metricValue interface{}
	var ok bool
	switch {
	case value != nil:
		if metricValue, ok = value.(float64); !ok {
			return Metric{}, ErrIncorrectType
		}
	case delta != nil:
		if metricValue, ok = delta.(int64); !ok {
			return Metric{}, ErrIncorrectType
		}
	default:
		return Metric{}, ErrMetricNotFound
	}

	metric := Metric{
		Value: metricValue,
		Type:  MetricType(opts.MetricType),
	}
	return metric, nil
}

func (dbs *DBStorage) GetAll(ctx context.Context) (map[string]Metric, error) {
	rows, err := dbs.conn.Query(ctx, `SELECT name, type, value, delta FROM metrics`)
	if err != nil {
		return nil, fmt.Errorf("QueryContext error: %w", err)
	}
	defer rows.Close()

	metrics := make(map[string]Metric)
	for rows.Next() {
		var (
			name, t      string
			value, delta interface{}
		)
		if err := rows.Scan(&name, &t, &value, &delta); err != nil {
			fmt.Println(err)
			continue
		}

		var metricValue interface{}
		if value != nil {
			metricValue = value
		} else if delta != nil {
			metricValue = delta
		}

		metricKey := fmt.Sprintf("%s_%s", name, t)
		metrics[metricKey] = Metric{Type: MetricType(t), Value: metricValue}
	}

	return metrics, nil
}

func (dbs *DBStorage) SetAll(ctx context.Context, opts *SetAllOptions) error {
	for key, metric := range opts.Metrics {
		updateOpts := &UpdateOptions{
			MetricName: key,
			Update:     metric,
		}
		if err := dbs.Update(ctx, updateOpts); err != nil {
			return fmt.Errorf("can't update DBStorage by %s %s: %w", key, metric, err)
		}
	}

	return nil
}

func (dbs *DBStorage) Ping(ctx context.Context) error {
	if err := dbs.conn.Ping(ctx); err != nil {
		return fmt.Errorf("db ping error %w", err)
	}
	return nil
}

func (dbs *DBStorage) Close() error {
	dbs.conn.Close()
	return nil
}

func newDB(ctx context.Context, dataSourceName string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.Connect(ctx, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("can't open database %w", err)
	}
	return conn, nil
}

func (dbs *DBStorage) CreateTable(ctx context.Context) error { // TODO make with migrations
	query := `
  CREATE TABLE IF NOT EXISTS metrics (
   name text NOT NULL,
   type text NOT NULL,
   value double precision,
   delta bigint,
   PRIMARY KEY (name, type)
  );
 `

	_, err := dbs.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to create table %w", err)
	}

	return nil
}
