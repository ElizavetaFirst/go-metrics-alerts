package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
}

type memStorage struct {
	Data sync.Map
}

func NewMemStorage() *memStorage {
	return &memStorage{}
}

func (ms *memStorage) Update(opts *UpdateOptions) error {
	metricName := opts.MetricName
	update := opts.Update
	uniqueID := metricName + string(update.Type)
	m, exists := ms.Data.Load(uniqueID)
	if !exists {
		ms.Data.Store(uniqueID, update)
		return nil
	}
	metric, ok := m.(Metric)
	if !ok {
		return errors.New("can't get metric")
	}

	switch metric.Type {
	case Gauge:
		metric.Value = update.Value
	case Counter:
		if value, ok := metric.Value.(int64); ok {
			if newValue, ok := update.Value.(int64); ok {
				metric.Value = value + newValue
			}
		} else {
			return errors.New("unexpected value type for counter metric")
		}
	}

	ms.Data.Store(uniqueID, metric)
	return nil
}

func (ms *memStorage) Get(opts *GetOptions) (Metric, bool) {
	metricName := opts.MetricName
	metricType := opts.MetricType
	uniqueID := metricName + metricType
	metric, exists := ms.Data.Load(uniqueID)
	if exists {
		return metric.(Metric), exists
	}
	return Metric{}, exists
}

func (ms *memStorage) GetAll(opts *GetAllOptions) map[string]Metric {
	result := make(map[string]Metric)
	ms.Data.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			fmt.Printf("can't get key value")
		}

		valueMetric, ok := value.(Metric)
		if !ok {
			fmt.Printf("can't get value")
		}

		result[keyStr] = valueMetric
		return true
	})
	return result
}

func (ms *memStorage) SetAll(opts *SetAllOptions) {
	metrics := opts.Metrics
	for key, metric := range metrics {
		if metric.Type == "counter" && metric.Value != nil {
			if value, ok := metric.Value.(float64); ok {
				metric.Value = int64(value)
			}
		}
		ms.Data.Store(key, metric)
	}
}
