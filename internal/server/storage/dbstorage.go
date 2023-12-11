package storage

import (
	"context"
	"fmt"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/db"
	"github.com/pkg/errors"
)

type DBStorage struct {
	db *db.DB
}

func NewPostgresStorage(ctx context.Context, databaseDSN string) (*DBStorage, error) {
	realDb, err := db.NewDB(ctx, databaseDSN)
	if err != nil {
		return nil, ErrDBNotInited
	}

	err = realDb.CreateTable(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't create table")
	}

	return &DBStorage{
		db: realDb,
	}, nil
}

func (dbs *DBStorage) Update(ctx context.Context, opts *UpdateOptions) error {
	if dbs.db == nil {
		return errors.New("database is not inited")
	}
	_, err := dbs.db.ExecContext(ctx, `INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3) 
 ON CONFLICT(name, type) DO UPDATE SET value = $3;`, opts.MetricName, opts.Update.Type, opts.Update.Value)
	return err
}

func (dbs *DBStorage) Get(ctx context.Context, opts *GetOptions) (Metric, error) {
	if dbs.db == nil {
		return Metric{}, fmt.Errorf("database is not inited: %w", ErrDBNotInited)
	}

	row := dbs.db.QueryRowContext(ctx, `SELECT value FROM metrics WHERE name=$1 AND type=$2`, opts.MetricName, opts.MetricType)

	var value int
	err := row.Scan(&value)
	if err != nil {
		fmt.Println(err)
		return Metric{}, fmt.Errorf("can't get metric from MemStorage %s %s: %w", opts.MetricName, opts.MetricType, ErrMetricNotFound)
	}

	metric := Metric{
		Value: value,
		Type:  MetricType(opts.MetricType),
	}
	return metric, nil
}

func (dbs *DBStorage) GetAll(ctx context.Context) (map[string]Metric, error) {
	if dbs.db == nil {
		return nil, fmt.Errorf("database is not inited: %w", ErrDBNotInited)
	}
	rows, err := dbs.db.QueryContext(ctx, `SELECT name, type, value FROM metrics`)
	if err != nil {
		return nil, fmt.Errorf("QueryContext error: %w", err)
	}
	defer rows.Close()

	metrics := make(map[string]Metric)
	for rows.Next() {
		var (
			name  string
			t     string
			value float64
		)
		if err := rows.Scan(&name, &t, &value); err != nil {
			fmt.Println(err) // handle error properly
			continue
		}
		metricKey := fmt.Sprintf("%s_%s", name, t)
		metrics[metricKey] = Metric{Type: MetricType(t), Value: value}
	}

	return metrics, nil
}

func (dbs *DBStorage) SetAll(ctx context.Context, opts *SetAllOptions) error {
	if dbs.db == nil {
		return fmt.Errorf("database is not inited: %w", ErrDBNotInited)
	}
	for key, metric := range opts.Metrics {
		updateOpts := &UpdateOptions{
			MetricName: key,
			Update:     metric,
		}
		if err := dbs.Update(ctx, updateOpts); err != nil {
			return fmt.Errorf("can't update DBStorage by %s %s: %w", key, metric, err) // you would want to handle error properly
		}
	}
	return nil
}

func (dbs *DBStorage) Ping() error {
	if dbs.db == nil {
		return errors.New("database is not inited")
	}
	return dbs.db.Ping()
}

func (dbs *DBStorage) Close() error {
	if dbs.db == nil {
		return errors.New("database is not inited")
	}
	return dbs.db.Close()
}
