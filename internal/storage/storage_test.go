package storage

import (
	"testing"
)

func TestGetGauge(t *testing.T) {
	ms := NewMemStorage()

	_, ok := ms.Get("nonexistent")
	if ok == true {
		t.Errorf("expected false for nonexistent metric, got %v", ok)
	}

	ms.Update("testMetric", Metric{
		Type:  Gauge,
		Value: 23.5,
	})

	metric, ok := ms.Get("testMetric")
	if ok == false {
		t.Errorf("expected true for existent metric, got %v", ok)
	} else if metric.Value != 23.5 {
		t.Errorf("expected 23.5 for existent metric, got %v", metric.Value)
	}
}

func TestGetCounter(t *testing.T) {
	ms := NewMemStorage()

	_, ok := ms.Get("nonexistent")
	if ok == true {
		t.Errorf("expected false for nonexistent metric, got %v", ok)
	}

	ms.Update("testMetric", Metric{
		Type:  Counter,
		Value: 5,
	})

	ms.Update("testMetric", Metric{
		Type:  Counter,
		Value: 5,
	})

	metric, ok := ms.Get("testMetric")
	if ok == false {
		t.Errorf("expected true for existent metric, got %v", ok)
	} else if metric.Value != 10 {
		t.Errorf("expected 10 for existent metric, got %v", metric.Value)
	}
}