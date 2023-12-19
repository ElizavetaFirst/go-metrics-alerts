package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
)

type MemStorage struct {
	Data sync.Map
}

func NewMemStorage() *MemStorage {
	return &MemStorage{}
}

func (ms *MemStorage) Update(ctx context.Context, opts *UpdateOptions) error {
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

func (ms *MemStorage) Get(ctx context.Context, opts *GetOptions) (Metric, error) {
	metricName := opts.MetricName
	metricType := opts.MetricType
	uniqueID := metricName + metricType
	metric, exists := ms.Data.Load(uniqueID)
	if exists {
		return metric.(Metric), nil
	}
	return Metric{}, fmt.Errorf("can't get metric from MemStorage %s %s: %w", metricName, metricType, ErrMetricNotFound)
}

func (ms *MemStorage) GetAll(ctx context.Context) (map[string]Metric, error) {
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
	return result, nil
}

func (ms *MemStorage) SetAll(ctx context.Context, opts *SetAllOptions) error {
	metrics := opts.Metrics
	for key, metric := range metrics {
		if metric.Type == constants.Counter && metric.Value != nil {
			if value, ok := metric.Value.(float64); ok {
				metric.Value = int64(value)
			}
		}
		ms.Data.Store(key, metric)
	}
	return nil
}

func (ms *MemStorage) Ping() error {
	return nil
}

func (ms *MemStorage) Close() error {
	return nil
}
