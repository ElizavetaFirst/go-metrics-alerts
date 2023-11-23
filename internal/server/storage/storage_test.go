package storage

import (
	"fmt"
	"testing"
)

func TestGetGauge(t *testing.T) {
	ms := NewMemStorage()

	_, ok := ms.Get("nonexistent", "gauge")
	if ok == true {
		t.Errorf("expected false for nonexistent metric, got %v", ok)
	}

	err := ms.Update("testMetric", Metric{
		Type:  Gauge,
		Value: 23.5,
	})
	if err != nil {
		fmt.Printf("can't update testMetric %v", err)
	}

	metric, ok := ms.Get("testMetric", "gauge")
	if ok == false {
		t.Errorf("expected true for existent metric, got %v", ok)
	} else if metric.Value != 23.5 {
		t.Errorf("expected 23.5 for existent metric, got %v", metric.Value)
	}
}

func TestGetCounter(t *testing.T) {
	ms := NewMemStorage()

	_, ok := ms.Get("nonexistent", "counter")
	if ok == true {
		t.Errorf("expected false for nonexistent metric, got %v", ok)
	}

	err := ms.Update("testMetric", Metric{
		Type:  Counter,
		Value: int64(5),
	})
	if err != nil {
		fmt.Printf("can't update testMetric %v", err)
	}

	err = ms.Update("testMetric", Metric{
		Type:  Counter,
		Value: int64(5),
	})
	if err != nil {
		fmt.Printf("can't update testMetric %v", err)
	}

	metric, ok := ms.Get("testMetric", "counter")
	if ok == false {
		t.Errorf("expected true for existent metric, got %v", ok)
	} else if metric.Value != int64(10) {
		t.Errorf("expected 10 for existent metric, got %v", metric.Value)
	}
}
