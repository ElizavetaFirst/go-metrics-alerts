package storage

import (
	"context"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/pkg/errors"
)

const (
	Gauge   MetricType = constants.Gauge
	Counter MetricType = constants.Counter
)

type (
	MetricType string

	UpdateOptions struct {
		MetricName string
		Update     Metric
	}

	GetOptions struct {
		MetricName string
		MetricType string
	}

	SetAllOptions struct {
		Metrics map[string]Metric
	}
)

var (
	ErrDBNotInited    = errors.New("db is not inited")
	ErrMetricNotFound = errors.New("metric not found")
	ErrIncorrectType  = errors.New("incorrect metric type")
)

type Storage interface {
	Update(ctx context.Context, opts *UpdateOptions) error
	Get(ctx context.Context, opts *GetOptions) (Metric, error)
	GetAll(ctx context.Context) (map[string]Metric, error)
	SetAll(ctx context.Context, opts *SetAllOptions) error
	Ping() error
	Close() error
}
