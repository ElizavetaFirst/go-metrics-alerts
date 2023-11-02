package uploader

import (
	"net/http"
	"strconv"
	"time"
)

const (
	reportInterval = 10 * time.Second
	baseURL        = "http://localhost:8080"
)

type Uploader struct {
	counterMetricsFunc func() map[string]int64
	gaugeMetricsFunc   func() map[string]float64
}

func NewUploader(gaugeMetricsFunc func() map[string]float64, counterMetricsFunc func() map[string]int64) *Uploader {
	return &Uploader{gaugeMetricsFunc: gaugeMetricsFunc, counterMetricsFunc: counterMetricsFunc}
}

func (u *Uploader) Run() {
	ticker := time.NewTicker(reportInterval)
	for {
		<-ticker.C
		u.SendGaugeMetrics(u.gaugeMetricsFunc())
		u.SendCounterMetrics(u.counterMetricsFunc())
	}
}

func (u *Uploader) SendGaugeMetrics(metrics map[string]float64) {
	client := &http.Client{}
	for k, v := range metrics {
		req, _ := http.NewRequest("POST", baseURL+"/update/gauge/"+k+"/"+strconv.FormatFloat(v, 'f', -1, 64), nil)
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
		req, _ := http.NewRequest("POST", baseURL+"/update/counter/"+k+"/"+strconv.FormatInt(v, 10), nil)
		req.Header.Set("Content-Type", "text/plain")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
	}
}
