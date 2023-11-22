package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/metrics"
	"github.com/pkg/errors"
)

const (
	contentTypeStr = "Content-Type"
	textPlainStr   = "text/plain"
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
	for range ticker.C {
		if err := u.SendGaugeMetrics(u.gaugeMetricsFunc()); err != nil {
			u.errorChan <- err
			return
		}
		if err := u.SendCounterMetrics(u.counterMetricsFunc()); err != nil {
			u.errorChan <- err
			return
		}
		/*if err := u.SendGaugeMetricsJson(u.gaugeMetricsFunc()); err != nil {
			u.errorChan <- err
			return
		}
		if err := u.SendCounterMetricsJson(u.counterMetricsFunc()); err != nil {
			u.errorChan <- err
			return
		}*/
	}
}

func (u *Uploader) sendMetrics(url string) error {
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set(contentTypeStr, textPlainStr)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "can't send update request")
	}
	if err = resp.Body.Close(); err != nil {
		return errors.Wrap(err, "can't close update request resp.Body")
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

func (u *Uploader) sendMetricsJson(url string, metrics metrics.Metrics) error {
	client := &http.Client{}

	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return errors.Wrap(err, "can't marshal metrics to JSON")
	}

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(metricsJSON))
	req.Header.Set(contentTypeStr, "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "can't send update request")
	}
	if err = resp.Body.Close(); err != nil {
		return errors.Wrap(err, "can't close update request resp.Body")
	}
	return nil
}

func (u *Uploader) SendGaugeMetricsJson(metricsMap map[string]float64) error {
	for k, v := range metricsMap {
		metric := metrics.Metrics{
			ID:    k,
			MType: "gauge",
			Value: &v,
		}

		url := fmt.Sprintf("http://%s/update/gauge/%s/%f", u.addr, k, v)
		if err := u.sendMetricsJson(url, metric); err != nil {
			return err
		}
	}
	return nil
}

func (u *Uploader) SendCounterMetricsJson(metricsMap map[string]int64) error {
	for k, v := range metricsMap {
		metric := metrics.Metrics{
			ID:    k,
			MType: "counter",
			Delta: &v,
		}

		url := fmt.Sprintf("http://%s/update/counter/%s/%d", u.addr, k, v)
		if err := u.sendMetricsJson(url, metric); err != nil {
			return err
		}
	}
	return nil
}
