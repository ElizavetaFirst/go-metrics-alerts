package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
)

type MemStorage struct {
	data sync.Map
}

func NewMemStorage() *MemStorage {
	return &MemStorage{}
}

func (ms *MemStorage) Update(ctx context.Context, opts *UpdateOptions) error {
	metricName := opts.MetricName
	update := opts.Update
	uniqueID := metricName + string(update.Type)
	m, exists := ms.data.Load(uniqueID)
	if !exists {
		ms.data.Store(uniqueID, update)
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
			ctx.Value(constants.Logger).(*zap.Logger).Error("unexpected value type for counter metric",
				zap.String("uniqueID", uniqueID))
			return errors.New("unexpected value type for counter metric")
		}
	}

	ms.data.Store(uniqueID, metric)
	return nil
}

func (ms *MemStorage) Get(ctx context.Context, opts *GetOptions) (Metric, error) {
	metricName := opts.MetricName
	metricType := opts.MetricType
	uniqueID := metricName + metricType
	metric, exists := ms.data.Load(uniqueID)
	if exists {
		return metric.(Metric), nil
	}
	ctx.Value(constants.Logger).(*zap.Logger).Error("can't get metric from MemStorage",
		zap.String("MetricName", metricName),
		zap.String("MetricType", metricType))
	return Metric{}, fmt.Errorf("can't get metric from MemStorage %s %s: %w", metricName, metricType, ErrMetricNotFound)
}

func (ms *MemStorage) GetAll(ctx context.Context) (map[string]Metric, error) {
	result := make(map[string]Metric)
	ms.data.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			ctx.Value(constants.Logger).(*zap.Logger).Warn("can't get key value")
		}

		valueMetric, ok := value.(Metric)
		if !ok {
			ctx.Value(constants.Logger).(*zap.Logger).Warn("can't get value")
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
		ms.data.Store(key, metric)
	}
	return nil
}

func (ms *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (ms *MemStorage) Close() error {
	return nil
}
