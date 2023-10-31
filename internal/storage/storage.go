package storage

import "sync"

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Metric struct {
	Type  MetricType
	Value float64
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
	if metric.Type == Counter {
		ms.Data[metricName] = Metric{Type: Counter, Value: metric.Value + update.Value}
	} else {
		ms.Data[metricName] = update
	}
	return nil
}

func (ms *MemStorage) Get(metricName string) (Metric, bool) {
	ms.RLock()
	defer ms.RUnlock()
	metric, exists := ms.Data[metricName]
	return metric, exists
}
