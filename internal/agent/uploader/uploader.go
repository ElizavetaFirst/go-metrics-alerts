package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/metrics"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	contentTypeStr = "Content-Type"
	textPlainStr   = "text/plain"
	maxTimeout     = 30
)

type (
	GaugeMetricsFuncType   func() map[string]float64
	CounterMetricsFuncType func() map[string]int64
	Uploader               struct {
		counterMetricsFunc CounterMetricsFuncType
		gaugeMetricsFunc   GaugeMetricsFuncType
		errorChan          chan error
		addr               string
		reportInterval     time.Duration
	}
)

func NewUploader(
	addr string, reportInterval time.Duration,
	gaugeMetricsFunc GaugeMetricsFuncType,
	counterMetricsFunc CounterMetricsFuncType,
	errorChan chan error,
) *Uploader {
	return &Uploader{
		gaugeMetricsFunc:   gaugeMetricsFunc,
		counterMetricsFunc: counterMetricsFunc,
		addr:               addr,
		reportInterval:     reportInterval,
		errorChan:          errorChan,
	}
}

func (u *Uploader) Run() {
	ticker := time.NewTicker(u.reportInterval)

	errorCount := 0
	for range ticker.C {
		for {
			if err := u.SendGaugeMetricsJSON(u.gaugeMetricsFunc()); err != nil {
				fmt.Printf("SendGaugeMetricsJson return error %v", err)
				errorCount++
				if errorCount >= constants.MaxErrors {
					u.errorChan <- err
					return
				}
				continue
			}
			if err := u.SendCounterMetricsJSON(u.counterMetricsFunc()); err != nil {
				fmt.Printf("SendCounterMetricsJson return error %v", err)
				errorCount++
				if errorCount >= constants.MaxErrors {
					u.errorChan <- err
					return
				}
				continue
			}
			break
		}
	}
}

func (u *Uploader) createRetryableHTTPClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = 4
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 5 * time.Second
	return client
}

func (u *Uploader) sendMetrics(url string) error {
	client := u.createRetryableHTTPClient()
	req, err := retryablehttp.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("can't make request %w", err)
	}
	req.Header.Set(contentTypeStr, textPlainStr)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("can't send update request %w", err)
	}
	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("can't close update request resp.Body %w", err)
	}
	return nil
}

func (u *Uploader) SendGaugeMetrics(metrics map[string]float64) error {
	for k, v := range metrics {
		url := fmt.Sprintf("http://%s/update/gauge/%s/%f", u.addr, k, v)
		if err := u.sendMetrics(url); err != nil {
			return err
		}
	}
	return nil
}

func (u *Uploader) SendCounterMetrics(metrics map[string]int64) error {
	for k, v := range metrics {
		url := fmt.Sprintf("http://%s/update/counter/%s/%d", u.addr, k, v)
		if err := u.sendMetrics(url); err != nil {
			return err
		}
	}
	return nil
}

func (u *Uploader) sendMetricsJSON(url string, metrics []byte) error {
	retryableClient := u.createRetryableHTTPClient()
	client := &ClientWithMiddleware{
		HTTPClient: retryableClient,
	}
	req, err := retryablehttp.NewRequest(http.MethodPost, url, bytes.NewBuffer(metrics))
	if err != nil {
		return fmt.Errorf("can't make request %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("can't send update request %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			closeErr = fmt.Errorf("failed to close the body of the response %w", closeErr)
			if err == nil {
				err = closeErr
			}
		}
	}()
	return nil
}

func (u *Uploader) SendGaugeMetricsUpdatesJSON(metricsMap map[string]float64) error {
	metricsList := make([]metrics.Metrics, 0, len(metricsMap))
	for k, v := range metricsMap {
		v := v
		metric := metrics.Metrics{
			ID:    k,
			MType: constants.Gauge,
			Value: &v,
		}
		metricsList = append(metricsList, metric)
	}

	url := fmt.Sprintf("http://%s/updates/", u.addr)
	metricsJSON, err := json.Marshal(metricsList)
	if err != nil {
		return fmt.Errorf("can't marshal metrics to JSON %w", err)
	}
	return u.sendMetricsJSON(url, metricsJSON)
}

func (u *Uploader) SendGaugeMetricsJSON(metricsMap map[string]float64) error {
	for k, v := range metricsMap {
		v := v
		metric := metrics.Metrics{
			ID:    k,
			MType: constants.Gauge,
			Value: &v,
		}

		url := fmt.Sprintf("http://%s/update", u.addr)
		metricsJSON, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("can't marshal metrics to JSON %w", err)
		}
		if err := u.sendMetricsJSON(url, metricsJSON); err != nil {
			return err
		}
	}
	return nil
}

func (u *Uploader) SendCounterMetricsJSON(metricsMap map[string]int64) error {
	for k, v := range metricsMap {
		v := v
		metric := metrics.Metrics{
			ID:    k,
			MType: constants.Counter,
			Delta: &v,
		}

		metricsJSON, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("can't marshal metrics to JSON %w", err)
		}
		url := fmt.Sprintf("http://%s/update", u.addr)
		if err := u.sendMetricsJSON(url, metricsJSON); err != nil {
			return err
		}
	}
	return nil
}

func (u *Uploader) SendCounterMetricsUpdatesJSON(metricsMap map[string]int64) error {
	metricsList := make([]metrics.Metrics, 0, len(metricsMap))
	for k, v := range metricsMap {
		v := v
		metric := metrics.Metrics{
			ID:    k,
			MType: constants.Counter,
			Delta: &v,
		}
		metricsList = append(metricsList, metric)
	}

	url := fmt.Sprintf("http://%s/updates/", u.addr)
	metricsJSON, err := json.Marshal(metricsList)
	if err != nil {
		return fmt.Errorf("can't marshal metrics to JSON %w", err)
	}
	return u.sendMetricsJSON(url, metricsJSON)
}
