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

	var value, delta interface{}
	if opts.Update.Type == "counter" {
		delta = opts.Update.Value
	} else if opts.Update.Type == "gauge" {
		value = opts.Update.Value
	}

	_, err := dbs.db.ExecContext(ctx, `
		INSERT INTO metrics (name, type, value, delta)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(name, type) DO UPDATE
		SET value = EXCLUDED.value, delta = EXCLUDED.delta;`,
		opts.MetricName, opts.Update.Type, value, delta)
	return err
}

func (dbs *DBStorage) Get(ctx context.Context, opts *GetOptions) (Metric, error) {
	if dbs.db == nil {
		return Metric{}, fmt.Errorf("database is not inited: %w", ErrDBNotInited)
	}

	row := dbs.db.QueryRowContext(ctx, `SELECT value, delta FROM metrics WHERE name=$1 AND type=$2`, opts.MetricName, opts.MetricType)

	var value, delta interface{}
	err := row.Scan(&value, &delta)
	if err != nil {
		fmt.Println(err)
		return Metric{}, fmt.Errorf("can't get metric from MemStorage %s %s: %w", opts.MetricName, opts.MetricType, ErrMetricNotFound)
	}

	var metricValue interface{}
	if value != nil {
		metricValue = value
	} else if delta != nil {
		metricValue = delta.(int64)
	}

	metric := Metric{
		Value: metricValue,
		Type:  MetricType(opts.MetricType),
	}
	return metric, nil
}

func (dbs *DBStorage) GetAll(ctx context.Context) (map[string]Metric, error) {
	if dbs.db == nil {
		return nil, fmt.Errorf("database is not inited: %w", ErrDBNotInited)
	}

	rows, err := dbs.db.QueryContext(ctx, `SELECT name, type, value, delta FROM metrics`)
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
			metricValue = delta.(int64)
		}

		metricKey := fmt.Sprintf("%s_%s", name, t)
		metrics[metricKey] = Metric{Type: MetricType(t), Value: metricValue}
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
			return fmt.Errorf("can't update DBStorage by %s %s: %w", key, metric, err)
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
