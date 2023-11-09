package uploader

import (
	"net/http"
	"time"
	"fmt"
)

type Uploader struct {
	counterMetricsFunc func() map[string]int64
	gaugeMetricsFunc   func() map[string]float64
	addr               string
	reportInterval     time.Duration
}

func NewUploader(addr string, reportInterval time.Duration, gaugeMetricsFunc func() map[string]float64, counterMetricsFunc func() map[string]int64) *Uploader {
	return &Uploader{
		gaugeMetricsFunc:   gaugeMetricsFunc,
		counterMetricsFunc: counterMetricsFunc,
		addr:               addr,
		reportInterval:     reportInterval,
	}
}

func (u *Uploader) Run() {
	ticker := time.NewTicker(u.reportInterval)

	for {
		<-ticker.C
		u.SendGaugeMetrics(u.gaugeMetricsFunc())
		u.SendCounterMetrics(u.counterMetricsFunc())
	}
}

func (u *Uploader) SendGaugeMetrics(metrics map[string]float64) {
	client := &http.Client{}
	for k, v := range metrics {
		req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/update/gauge/%s/%f", u.addr, k, v), nil)
		req.Header.Set("Content-Type", "text/plain")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
	}
}

func (u *Uploader) SendCounterMetrics(metrics map[string]int64) {
	client := &http.Client{}
	for k, v := range metrics {
		req, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/update/counter/%s/%d", u.addr, k, v), nil)
		req.Header.Set("Content-Type", "text/plain")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
	}
}
