package storage

import (
	"errors"
	"sync"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Storage interface {
	Update(metricName string, update Metric) error
	Get(metricName string) (Metric, bool)
	GetAll() map[string]Metric
}

type memStorage struct {
	Data sync.Map
}

func NewMemStorage() *memStorage {
	return &memStorage{}
}

func (ms *memStorage) Update(metricName string, update Metric) error {
	m, exists := ms.Data.Load(metricName)
	if !exists {
		ms.Data.Store(metricName, update)
		return nil
	}
	metric := m.(Metric)
	if metric.Type != update.Type {
		return nil
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

	ms.Data.Store(metricName, metric)
	return nil
}

func (ms *memStorage) Get(metricName string) (Metric, bool) {
	metric, exists := ms.Data.Load(metricName)
	if exists {
		return metric.(Metric), exists
	}
	return Metric{}, exists
}

func (ms *memStorage) GetAll() map[string]Metric {
	result := make(map[string]Metric)
	ms.Data.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(Metric)
		return true
	})
	return result
}
