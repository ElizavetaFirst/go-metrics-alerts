package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
)

type MetricType string

const (
	Gauge   MetricType = constants.Gauge
	Counter MetricType = constants.Counter
)

type Storage interface {
	Update(metricName string, update Metric) error
	Get(metricName string, metricType string) (Metric, bool)
	GetAll() map[string]Metric
}

type memStorage struct {
	Data sync.Map
}

func NewMemStorage() *memStorage {
	return &memStorage{}
}

func (ms *memStorage) Update(metricName string, update Metric) error {
	uniqueID := metricName + string(update.Type)
	fmt.Println(uniqueID)
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

func (ms *memStorage) Get(metricName string, metricType string) (Metric, bool) {
	uniqueID := metricName + metricType
	metric, exists := ms.Data.Load(uniqueID)
	if exists {
		return metric.(Metric), exists
	}
	return Metric{}, exists
}

func (ms *memStorage) GetAll() map[string]Metric {
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
