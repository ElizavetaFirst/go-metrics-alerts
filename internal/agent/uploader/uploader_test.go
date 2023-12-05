package uploader

import (
	"fmt"
	"io/ioutil"
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

func TestUploader_SendGaugeMetricsJson(t *testing.T) {
	respRecorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed reading request body: %v", err)
		}
		if string(reqBody) != "{\"ID\":\"metric1\",\"MType\":\"Gauge\",\"Value\":0.1}" {
			t.Fatalf("Expected JSON data did not match actual data")
		}
	})

	handler.ServeHTTP(respRecorder, req)

	if status := respRecorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"ID":"metric1","MType":"Gauge","Value":0.1}`
	if respRecorder.Body.String() != "{\"ID\":\"metric1\",\"MType\":\"Gauge\",\"Value\":0.1}" {
		t.Errorf("Handler returned unexpected body: got %v want %v", respRecorder.Body.String(), expected)
	}
}
