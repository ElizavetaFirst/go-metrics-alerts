package storage

import (
	"database/sql"
	"fmt"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/db"
	"github.com/pkg/errors"
)

type PostgresStorage struct {
	db db.SQLDB
}

func NewPostgresStorage(databaseDSN string) (*PostgresStorage, error) {
	db, err := db.NewDB(databaseDSN)
	if err != nil {
		return nil, errors.Wrap(err, "can't init db")
	}
	return &PostgresStorage{
		db: db,
	}, nil
}

func (ps *PostgresStorage) Update(opts *UpdateOptions) error {
	_, err := ps.db.ExecContext(opts.Context, `INSERT INTO metrics (name, type, value) VALUES ($1, $2, $3) 
	 ON CONFLICT(name, type) DO UPDATE SET value = $3;`, opts.MetricName, opts.Update.Type, opts.Update.Value)
	return err
}

func (ps *PostgresStorage) Get(opts *GetOptions) (Metric, bool) {
	fmt.Println(ps.db)
	row := ps.db.QueryRowContext(opts.Context, `SELECT value FROM metrics WHERE name=$1 AND type=$2`, opts.MetricName, opts.MetricType)

	var value int
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		return Metric{}, false
	}
	if err != nil {
		fmt.Println(err) // handle error properly
		return Metric{}, false
	}

	metric := Metric{
		Value: value,
		Type:  MetricType(opts.MetricType),
	}
	return metric, true
}

func (ps *PostgresStorage) GetAll(opts *GetAllOptions) map[string]Metric {
	rows, err := ps.db.QueryContext(opts.Context, `SELECT name, type, value FROM metrics`)
	if err != nil {
		fmt.Println(err) // handle error properly
		return nil
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

	return metrics
}

func (ps *PostgresStorage) SetAll(opts *SetAllOptions) {
	for key, metric := range opts.Metrics {
		updateOpts := &UpdateOptions{
			Context:    opts.Context,
			MetricName: key,
			Update:     metric,
		}
		if err := ps.Update(updateOpts); err != nil {
			fmt.Println(err) // you would want to handle error properly
		}
	}
}

func (ps *PostgresStorage) Ping() error {
	return ps.db.Ping()
}

func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}
