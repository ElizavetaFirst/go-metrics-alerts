package storage

import "sync"

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Metric struct {
	Type  MetricType
	Value any
}

type Storage interface {
	Update(metricName string, update Metric) error
	Get(metricName string) (Metric, bool)
}

type MemStorage struct {
	Data map[string]Metric
	sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Data: make(map[string]Metric)}
}

func (ms *MemStorage) Update(metricName string, update Metric) error {
	ms.Lock()
	defer ms.Unlock()
	metric, exists := ms.Data[metricName]
	if !exists {
		ms.Data[metricName] = update
		return nil
	}
	if metric.Type != update.Type {
		return nil
	}

	switch newValue := update.Value.(type) {
	case float64:
		if metric.Type == Counter {
			metric.Value = metric.Value.(float64) + newValue
		} else {
			metric.Value = newValue
		}
	case int64:
		if metric.Type == Counter {
			metric.Value = metric.Value.(int64) + newValue
		} else {
			metric.Value = newValue
		}
	}
	ms.Data[metricName] = metric
	return nil
}

func (ms *MemStorage) Get(metricName string) (Metric, bool) {
	ms.RLock()
	defer ms.RUnlock()
	metric, exists := ms.Data[metricName]
	return metric, exists
}
