package storage

import (
	"context"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
)

type MetricType string

type UpdateOptions struct {
	//nolint:containedctx // need for compact representation
	Context    context.Context
	MetricName string
	Update     Metric
}

type GetOptions struct {
	//nolint:containedctx // need for compact representation
	Context    context.Context
	MetricName string
	MetricType string
}

type GetAllOptions struct {
	//nolint:containedctx // need for compact representation
	Context context.Context
}

type SetAllOptions struct {
	//nolint:containedctx // need for compact representation
	Context context.Context
	Metrics map[string]Metric
}

const (
	Gauge   MetricType = constants.Gauge
	Counter MetricType = constants.Counter
)

type Storage interface {
	Update(opts *UpdateOptions) error
	Get(opts *GetOptions) (Metric, bool)
	GetAll(opts *GetAllOptions) map[string]Metric
	SetAll(opts *SetAllOptions)
	Ping() error
	Close() error
}
