package collector

import (
	"reflect"
	"testing"
	"time"
)

func TestCollector_GetGaugeMetrics(t *testing.T) {
	errorChan := make(chan error)
	collector := NewCollector(10*time.Second, errorChan)
	collector.GaugeMetrics = map[string]float64{
		"Alloc": 10.0,
	}

	got := collector.GetGaugeMetrics()

	if !reflect.DeepEqual(got, collector.GaugeMetrics) {
		t.Errorf("Expected %+v, Got %+v", collector.GaugeMetrics, got)
	}
}

func TestCollector_GetCounterMetrics(t *testing.T) {
	errorChan := make(chan error)
	collector := NewCollector(10*time.Second, errorChan)
	collector.CounterMetrics = map[string]int64{
		"Alloc": 5,
	}

	got := collector.GetCounterMetrics()

	if !reflect.DeepEqual(got, collector.CounterMetrics) {
		t.Errorf("Expected %+v, Got %+v", collector.CounterMetrics, got)
	}
}

func TestNewCollector(t *testing.T) {
	errorChan := make(chan error)
	c := NewCollector(10*time.Second, errorChan)
	if c.pollInterval != 10*time.Second {
		t.Errorf("Expected poll interval to be 10s, but got %v", c.pollInterval)
	}
	if c.errorChan != errorChan {
		t.Errorf("Expected error channel to be %v, but got %v", errorChan, c.errorChan)
	}
}
