package saver

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
)

func Test_saveMetricsToFile(t *testing.T) {
	filePath := "/tmp/metrics-db.json"

	metrics := make(map[string]storage.Metric)
	metrics["test_metric"] = storage.Metric{
		Type:  constants.Gauge,
		Value: 1.0,
	}

	err := saveMetricsToFile(metrics, filePath)
	if err != nil {
		t.Fatal(err)
	}

	loadedMetrics, err := loadMetricsFromFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	metric, ok := loadedMetrics["test_metric"]
	if !ok {
		t.Fatal("Test metric not found")
	}

	if metric.Type != constants.Gauge || metric.Value != 1.0 {
		t.Fatalf("Incorrect metric value; got %+v", metric)
	}

	if err := os.Remove(filePath); err != nil {
		t.Fatal(err)
	}
}

func Test_loadMetricsFromFile(t *testing.T) {
	filePath := "/tmp/metrics-db.json"
	metrics := make(map[string]storage.Metric)
	metrics["test_metric"] = storage.Metric{
		Type:  constants.Gauge,
		Value: 1.0,
	}

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(metrics); err != nil {
		t.Fatal(err)
	}

	if err = file.Close(); err != nil {
		t.Fatal(err)
	}

	loadedMetrics, err := loadMetricsFromFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	metric, ok := loadedMetrics["test_metric"]
	if !ok {
		t.Fatal("Test metric not found")
	}

	if metric.Type != constants.Gauge || metric.Value != 1.0 {
		t.Fatalf("Incorrect metric value; got %+v", metric)
	}

	if err := os.Remove(filePath); err != nil {
		t.Fatal(err)
	}
}
