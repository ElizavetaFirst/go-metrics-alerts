package storage

import (
	"fmt"
	"testing"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
)

func TestGetGauge(t *testing.T) {
	ms := NewMemStorage()

	_, ok := ms.Get(&GetOptions{MetricName: "nonexistent", MetricType: constants.Gauge})
	if ok == true {
		t.Errorf("expected false for nonexistent metric, got %v", ok)
	}

	err := ms.Update(&UpdateOptions{
		MetricName: "testMetric",
		Update: Metric{
			Type:  Gauge,
			Value: 23.5,
		},
	})
	if err != nil {
		fmt.Printf("can't update testMetric %v", err)
	}

	metric, ok := ms.Get(&GetOptions{MetricName: "testMetric", MetricType: constants.Gauge})
	if ok == false {
		t.Errorf("expected true for existent metric, got %v", ok)
	} else if metric.Value != 23.5 {
		t.Errorf("expected 23.5 for existent metric, got %v", metric.Value)
	}
}

func TestGetCounter(t *testing.T) {
	ms := NewMemStorage()

	_, ok := ms.Get(&GetOptions{MetricName: "nonexistent", MetricType: constants.Counter})
	if ok == true {
		t.Errorf("expected false for nonexistent metric, got %v", ok)
	}

	err := ms.Update(&UpdateOptions{
		MetricName: "testMetric",
		Update: Metric{
			Type:  Counter,
			Value: int64(10),
		},
	})
	if err != nil {
		fmt.Printf("can't update testMetric %v", err)
	}

	metric, ok := ms.Get(&GetOptions{MetricName: "testMetric", MetricType: constants.Counter})
	if ok == false {
		t.Errorf("expected true for existent metric, got %v", ok)
	} else if metric.Value != int64(10) {
		t.Errorf("expected 10 for existent metric, got %v", metric.Value)
	}
}
