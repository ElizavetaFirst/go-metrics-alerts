package uploader

import (
	"fmt"
	"net/http"
	"time"
)

type (
	GaugeMetricsFuncType   func() map[string]float64
	CounterMetricsFuncType func() map[string]int64
	Uploader               struct {
		counterMetricsFunc CounterMetricsFuncType
		gaugeMetricsFunc   GaugeMetricsFuncType
		addr               string
		reportInterval     time.Duration
		errorChan          chan error
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
	time.NewTicker(u.reportInterval)

	ticker := time.NewTicker(u.reportInterval)
	for {
		select {
		case <-ticker.C:
			if err := u.SendGaugeMetrics(u.gaugeMetricsFunc()); err != nil {
				u.errorChan <- err
			}
			if err := u.SendCounterMetrics(u.counterMetricsFunc()); err != nil {
				u.errorChan <- err
			}
		}
	}
}

func (u *Uploader) SendGaugeMetrics(metrics map[string]float64) error {
	client := &http.Client{}
	for k, v := range metrics {
		req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/update/gauge/%s/%f", u.addr, k, v), nil)
		req.Header.Set("Content-Type", "text/plain")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}

func (u *Uploader) SendCounterMetrics(metrics map[string]int64) error {
	client := &http.Client{}
	for k, v := range metrics {
		req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/update/counter/%s/%d", u.addr, k, v), nil)
		req.Header.Set("Content-Type", "text/plain")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}
