package storage

import (
	"errors"
	"fmt"
	"sync"
)

type MemStorage struct {
	Data sync.Map
}

func NewMemStorage() *MemStorage {
	return &MemStorage{}
}

func (ms *MemStorage) Update(opts *UpdateOptions) error {
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

func (ms *MemStorage) Get(opts *GetOptions) (Metric, bool) {
	metricName := opts.MetricName
	metricType := opts.MetricType
	uniqueID := metricName + metricType
	metric, exists := ms.Data.Load(uniqueID)
	if exists {
		return metric.(Metric), exists
	}
	return Metric{}, exists
}

func (ms *MemStorage) GetAll(opts *GetAllOptions) map[string]Metric {
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

func (ms *MemStorage) SetAll(opts *SetAllOptions) {
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

func (ms *MemStorage) Ping() error {
	return nil
}

func (ms *MemStorage) Close() error {
	return nil
}
