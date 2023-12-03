package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/metrics"
	"github.com/pkg/errors"
)

const (
	contentTypeStr  = "Content-Type"
	textPlainStr    = "text/plain"
	maxTimeout      = 30
	cantSendUpdate  = "can't send update request"
	cantCloseBody   = "can't close update request resp.Body"
	updateReqFormat = "http://%s/update"
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

func (u *Uploader) sendMetrics(url string) error {
	client := &http.Client{
		Timeout: time.Second * maxTimeout,
	}
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set(contentTypeStr, textPlainStr)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, cantSendUpdate)
	}
	if err = resp.Body.Close(); err != nil {
		return errors.Wrap(err, cantCloseBody)
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
	client := &ClientWithMiddleware{
		HTTPClient: &http.Client{
			Timeout: time.Second * maxTimeout,
		},
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(metrics))
	if err != nil {
		return errors.Wrap(err, "can't make request")
	}
	_, err = client.Do(req)
	if err != nil {
		return errors.Wrap(err, cantSendUpdate)
	}
	return nil
}

func (u *Uploader) SendGaugeMetricsJSON(metricsMap map[string]float64) error {
	for k, v := range metricsMap {
		v := v
		metric := metrics.Metrics{
			ID:    k,
			MType: constants.Gauge,
			Value: &v,
		}

		url := fmt.Sprintf(updateReqFormat, u.addr)
		metricsJSON, err := json.Marshal(metric)
		if err != nil {
			return errors.Wrap(err, "can't marshal metrics to JSON")
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
			return errors.Wrap(err, "can't marshal metrics to JSON")
		}
		url := fmt.Sprintf(updateReqFormat, u.addr)
		if err := u.sendMetricsJSON(url, metricsJSON); err != nil {
			return err
		}
	}
	return nil
}
