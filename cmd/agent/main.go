package main

import (
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/collector"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/uploader"
)

func main() {
	c := collector.NewCollector()
	u := uploader.NewUploader(c.GetGaugeMetrics, c.GetCounterMetrics)

	go c.Run()
	u.Run()
}

