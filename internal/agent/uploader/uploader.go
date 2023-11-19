package uploader

import (
	"fmt"
	"net/http"
	"time"

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
	}
}

func (u *Uploader) sendMetrics(url string) error {
	client := &http.Client{}
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
