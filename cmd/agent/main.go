package main

import (
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/agent/collector"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/agent/uploader"
)

func main() {
	c := collector.NewCollector()
	u := uploader.NewUploader(c.GetGaugeMetrics, c.GetCounterMetrics)

	go c.Run()
	u.Run()
}
