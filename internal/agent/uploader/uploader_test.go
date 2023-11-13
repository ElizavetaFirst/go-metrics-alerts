package uploader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

var gaugeMetrics = func() map[string]float64 {
	return map[string]float64{
		"metric1": 0.1,
		"metric2": 0.2,
	}
}

var counterMetrics = func() map[string]int64 {
	return map[string]int64{
		"metric1": 1,
		"metric2": 2,
	}
}

func TestNewUploader(t *testing.T) {
	gaugeFunc := func() map[string]float64 {
		return nil
	}

	counterFunc := func() map[string]int64 {
		return nil
	}

	errorChan := make(chan error)
	uploader := NewUploader("localhost:8080", 2*time.Second, gaugeFunc, counterFunc, errorChan)

	if reflect.ValueOf(uploader.gaugeMetricsFunc).Pointer() != reflect.ValueOf(gaugeFunc).Pointer() {
		t.Error("Gauge metrics function not initialized correctly.")
	}

	if reflect.ValueOf(uploader.counterMetricsFunc).Pointer() != reflect.ValueOf(counterFunc).Pointer() {
		t.Error("Counter metrics function not initialized correctly.")
	}
}

func TestUploader_SendGaugeMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
	}))
	defer ts.Close()

	errorChan := make(chan error)
	uploader := NewUploader("localhost:8080", 2*time.Second, gaugeMetrics, counterMetrics, errorChan)

	if err := uploader.SendGaugeMetrics(gaugeMetrics()); err != nil {
		fmt.Printf("SendGaugeMetrics return error %v", err)
	}
}

func TestUploader_SendCounterMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
	}))
	defer ts.Close()

	errorChan := make(chan error)
	uploader := NewUploader("localhost:8080", 2*time.Second, gaugeMetrics, counterMetrics, errorChan)

	if err := uploader.SendCounterMetrics(counterMetrics()); err != nil {
		fmt.Printf("SendCounterMetrics return error %v", err)
	}
}
