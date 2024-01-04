package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(databaseDSN string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, databaseDSN)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

type DBStorage struct {
	conn *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, databaseDSN string) (*DBStorage, error) {
	if err := runMigrations(databaseDSN); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	conn, err := newDB(ctx, databaseDSN)
	if err != nil {
		ctx.Value(constants.Logger).(*zap.Logger).Error("database is not inited", zap.Error(err))
		return nil, fmt.Errorf("database is not inited: %w", ErrCantConnectDB)
	}

	dbStorage := DBStorage{
		conn: conn,
	}

	return &dbStorage, nil
}

func (dbs *DBStorage) Update(ctx context.Context, opts *UpdateOptions) error {
	var value, delta interface{}
	if opts.Update.Type == constants.Counter {
		delta = opts.Update.Value
	} else if opts.Update.Type == constants.Gauge {
		value = opts.Update.Value
	} else {
		ctx.Value(constants.Logger).(*zap.Logger).Error("incorrect type for update metric %s", zap.String("MetricType", string(opts.Update.Type)))
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
		ctx.Value(constants.Logger).(*zap.Logger).Error("ExecContext return error", zap.Error(err))
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
		ctx.Value(constants.Logger).(*zap.Logger).Error("can't get metric from DBStorage",
			zap.String("name", opts.MetricName),
			zap.String("type", opts.MetricType),
			zap.Error(err))
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
		ctx.Value(constants.Logger).(*zap.Logger).Error("QueryContext error", zap.Error(err))
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
			ctx.Value(constants.Logger).(*zap.Logger).Error("cant scan metric", zap.Error(err))
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
			ctx.Value(constants.Logger).(*zap.Logger).Error("can't update DBStorage by",
				zap.String("MetricName", key),
				zap.String("MetricType", string(metric.Type)),
				zap.Error(err))
			return fmt.Errorf("can't update DBStorage by %s %s: %w", key, metric, err)
		}
	}

	return nil
}

func (dbs *DBStorage) Ping(ctx context.Context) error {
	if err := dbs.conn.Ping(ctx); err != nil {
		ctx.Value(constants.Logger).(*zap.Logger).Error("db ping error", zap.Error(err))
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
		ctx.Value(constants.Logger).(*zap.Logger).Error("can't open database", zap.Error(err))
		return nil, fmt.Errorf("can't open database %w", err)
	}
	return conn, nil
}
