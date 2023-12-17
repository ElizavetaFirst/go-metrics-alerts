package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
)

//TODO Change later

func TestGetGauge(t *testing.T) {
	ms := NewMemStorage()

	ctx := context.TODO()
	_, err := ms.Get(ctx, &GetOptions{MetricName: "nonexistent", MetricType: constants.Gauge})
	//if err != nil {
	//	t.Errorf("expected false for nonexistent metric, got %v", err)
	//}

	err = ms.Update(ctx, &UpdateOptions{
		MetricName: "testMetric",
		Update: Metric{
			Type:  Gauge,
			Value: 23.5,
		},
	})
	if err != nil {
		fmt.Printf("can't update testMetric %v", err)
	}

	metric, err := ms.Get(ctx, &GetOptions{MetricName: "testMetric", MetricType: constants.Gauge})
	if err != nil {
		t.Errorf("expected true for existent metric, got %v", err)
	} else if metric.Value != 23.5 {
		t.Errorf("expected 23.5 for existent metric, got %v", metric.Value)
	}
}

func TestGetCounter(t *testing.T) {
	ms := NewMemStorage()

	ctx := context.TODO()
	_, err := ms.Get(ctx, &GetOptions{MetricName: "nonexistent", MetricType: constants.Counter})
	if err != nil {
		t.Errorf("expected false for nonexistent metric, got %v", err)
	}

	err = ms.Update(ctx, &UpdateOptions{
		MetricName: "testMetric",
		Update: Metric{
			Type:  Counter,
			Value: int64(10),
		},
	})
	if err != nil {
		fmt.Printf("can't update testMetric %v", err)
	}

	metric, err := ms.Get(ctx, &GetOptions{MetricName: "testMetric", MetricType: constants.Counter})
	if err != nil {
		t.Errorf("expected true for existent metric, got %v", err)
	} else if metric.Value != int64(10) {
		t.Errorf("expected 10 for existent metric, got %v", metric.Value)
	}
}
