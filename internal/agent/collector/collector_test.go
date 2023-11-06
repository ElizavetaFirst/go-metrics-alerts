package collector

import (
	"reflect"
	"testing"
	"time"
)

func TestCollector_GetGaugeMetrics(t *testing.T) {
	collector := NewCollector(10 * time.Second)
	collector.GaugeMetrics = map[string]float64{
		"Alloc": 10.0,
	}

	got := collector.GetGaugeMetrics()

	if !reflect.DeepEqual(got, collector.GaugeMetrics) {
		t.Errorf("Expected %+v, Got %+v", collector.GaugeMetrics, got)
	}
}

func TestCollector_GetCounterMetrics(t *testing.T) {
	collector := NewCollector(10 * time.Second)
	collector.CounterMetrics = map[string]int64{
		"Alloc": 5,
	}

	got := collector.GetCounterMetrics()

	if !reflect.DeepEqual(got, collector.CounterMetrics) {
		t.Errorf("Expected %+v, Got %+v", collector.CounterMetrics, got)
	}
}
