package uploader

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/metrics"
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
		log.Printf("SendGaugeMetrics return error %v", err)
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
		log.Printf("SendCounterMetrics return error %v", err)
	}
}

func newTestServer(t *testing.T, expectedMetric metrics.Metrics) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed reading request body: %v", err)
		}

		var metric metrics.Metrics
		if err := json.Unmarshal(reqBody, &metric); err != nil {
			t.Fatalf("Failed to unmarshal request body to Metrics struct: %v", err)
		}

		if !reflect.DeepEqual(metric, expectedMetric) {
			t.Fatalf("Metrics didn't match: got %v, expected: %v", metric, expectedMetric)
		}
	}))
}

func newTestUploader(t *testing.T, ts *httptest.Server) *Uploader {
	t.Helper()
	errorChan := make(chan error)
	trimmedURL := strings.TrimPrefix(ts.URL, "http://")
	return NewUploader(trimmedURL, 2*time.Second, gaugeMetrics, counterMetrics, errorChan)
}

//nolint:dupl // no way to delete duplicate
func TestUploader_SendGaugeMetricsJson(t *testing.T) {
	str := json.Number("0.1")
	f, err := str.Float64()
	if err != nil {
		t.Fatalf("Couldn't convert str to float64: %v", err)
	}

	expectedMetric := metrics.Metrics{
		ID:    "metric1",
		MType: constants.Gauge,
		Value: &f,
	}

	ts := newTestServer(t, expectedMetric)
	defer ts.Close()

	uploader := newTestUploader(t, ts)

	metricsMap := map[string]float64{
		"metric1": 0.1,
	}

	if err := uploader.SendGaugeMetricsJSON(metricsMap); err != nil {
		t.Fatalf("SendGaugeMetricsJson returned error: %v", err)
	}
}

//nolint:dupl // no way to delete duplicate
func TestUploader_SendCounterMetricsJson(t *testing.T) {
	str := json.Number("1")
	f, err := str.Int64()
	if err != nil {
		t.Fatalf("Couldn't convert str to int64: %v", err)
	}

	expectedMetric := metrics.Metrics{
		ID:    "metric1",
		MType: constants.Counter,
		Delta: &f,
	}

	ts := newTestServer(t, expectedMetric)
	defer ts.Close()

	uploader := newTestUploader(t, ts)

	metricsMap := map[string]int64{
		"metric1": 1,
	}

	if err := uploader.SendCounterMetricsJSON(metricsMap); err != nil {
		t.Fatalf("SendCounterMetricsJson returned error: %v", err)
	}
}
