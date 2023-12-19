package storage

import (
	"context"
	"fmt"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/db"
	"github.com/pkg/errors"
)

const databaseNotInitedFormat = "database is not inited: %w"

type DBStorage struct {
	db *db.DB
}

func NewPostgresStorage(ctx context.Context, databaseDSN string) (*DBStorage, error) {
	realDB, err := db.NewDB(ctx, databaseDSN)
	if err != nil {
		return nil, fmt.Errorf(databaseNotInitedFormat, ErrDBNotInited)
	}

	err = realDB.CreateTable(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't create table")
	}

	return &DBStorage{
		db: realDB,
	}, nil
}

func (dbs *DBStorage) Update(ctx context.Context, opts *UpdateOptions) error {
	if dbs.db == nil {
		return ErrDBNotInited
	}

	var value, delta interface{}
	if opts.Update.Type == constants.Counter {
		delta = opts.Update.Value
	} else if opts.Update.Type == constants.Gauge {
		value = opts.Update.Value
	}

	_, err := dbs.db.ExecContext(ctx, `
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
	if dbs.db == nil {
		return Metric{}, fmt.Errorf(databaseNotInitedFormat, ErrDBNotInited)
	}

	row := dbs.db.QueryRowContext(ctx, `SELECT value, delta FROM metrics WHERE name=$1 AND type=$2`,
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
		if metricValue, ok = value.(int64); !ok {
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
	if dbs.db == nil {
		return nil, fmt.Errorf(databaseNotInitedFormat, ErrDBNotInited)
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
			metricValue = delta
		}

		metricKey := fmt.Sprintf("%s_%s", name, t)
		metrics[metricKey] = Metric{Type: MetricType(t), Value: metricValue}
	}

	return metrics, nil
}

func (dbs *DBStorage) SetAll(ctx context.Context, opts *SetAllOptions) error {
	if dbs.db == nil {
		return fmt.Errorf(databaseNotInitedFormat, ErrDBNotInited)
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
		return fmt.Errorf(databaseNotInitedFormat, ErrDBNotInited)
	}
	if err := dbs.db.Ping(); err != nil {
		return fmt.Errorf("db Ping return error %w", err)
	}
	return nil
}

func (dbs *DBStorage) Close() error {
	if dbs.db == nil {
		return fmt.Errorf(databaseNotInitedFormat, ErrDBNotInited)
	}
	if err := dbs.db.Close(); err != nil {
		return fmt.Errorf("db Close return error %w", err)
	}
	return nil
}
