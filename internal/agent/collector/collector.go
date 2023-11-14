package collector

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

const randomFactor = 100

type Collector struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
	errorChan      chan error
	RandomValue    float64
	pollInterval   time.Duration
	PollCount      int64
}

func NewCollector(pollInterval time.Duration, errorChan chan error) *Collector {
	return &Collector{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
		pollInterval:   pollInterval,
		errorChan:      errorChan,
	}
}

func (c *Collector) Run() {
	ticker := time.NewTicker(c.pollInterval)
	var rtm runtime.MemStats
	for {
		select {
		case <-c.errorChan:
			fmt.Println("received error, stopping run")
			return
		default:
			<-ticker.C
			c.PollCount++
			c.RandomValue = rand.Float64()
			runtime.ReadMemStats(&rtm)
			c.updateMetrics(rtm)
		}
	}
}

func (c *Collector) updateMetrics(rtm runtime.MemStats) {
	c.GaugeMetrics = map[string]float64{
		"Alloc":         float64(rtm.Alloc),
		"BuckHashSys":   float64(rtm.BuckHashSys),
		"Frees":         float64(rtm.Frees),
		"GCCPUFraction": rtm.GCCPUFraction,
		"GCSys":         float64(rtm.GCSys),
		"HeapAlloc":     float64(rtm.HeapAlloc),
		"HeapIdle":      float64(rtm.HeapIdle),
		"HeapInuse":     float64(rtm.HeapInuse),
		"HeapObjects":   float64(rtm.HeapObjects),
		"HeapReleased":  float64(rtm.HeapReleased),
		"HeapSys":       float64(rtm.HeapSys),
		"LastGC":        float64(rtm.LastGC),
		"Lookups":       float64(rtm.Lookups),
		"MCacheInuse":   float64(rtm.MCacheInuse),
		"MCacheSys":     float64(rtm.MCacheSys),
		"MSpanInuse":    float64(rtm.MSpanInuse),
		"MSpanSys":      float64(rtm.MSpanSys),
		"Mallocs":       float64(rtm.Mallocs),
		"NextGC":        float64(rtm.NextGC),
		"NumForcedGC":   float64(rtm.NumForcedGC),
		"NumGC":         float64(rtm.NumGC),
		"OtherSys":      float64(rtm.OtherSys),
		"PauseTotalNs":  float64(rtm.PauseTotalNs),
		"StackInuse":    float64(rtm.StackInuse),
		"StackSys":      float64(rtm.StackSys),
		"Sys":           float64(rtm.Sys),
		"TotalAlloc":    float64(rtm.TotalAlloc),
		"PollCount":     float64(c.PollCount),
		"RandomValue":   c.RandomValue,
	}
	c.CounterMetrics = map[string]int64{
		"Alloc":         int64(rtm.Alloc),
		"BuckHashSys":   int64(rtm.BuckHashSys),
		"Frees":         int64(rtm.Frees),
		"GCCPUFraction": int64(rtm.GCCPUFraction),
		"GCSys":         int64(rtm.GCSys),
		"HeapAlloc":     int64(rtm.HeapAlloc),
		"HeapIdle":      int64(rtm.HeapIdle),
		"HeapInuse":     int64(rtm.HeapInuse),
		"HeapObjects":   int64(rtm.HeapObjects),
		"HeapReleased":  int64(rtm.HeapReleased),
		"HeapSys":       int64(rtm.HeapSys),
		"LastGC":        int64(rtm.LastGC),
		"Lookups":       int64(rtm.Lookups),
		"MCacheInuse":   int64(rtm.MCacheInuse),
		"MCacheSys":     int64(rtm.MCacheSys),
		"MSpanInuse":    int64(rtm.MSpanInuse),
		"MSpanSys":      int64(rtm.MSpanSys),
		"Mallocs":       int64(rtm.Mallocs),
		"NextGC":        int64(rtm.NextGC),
		"NumForcedGC":   int64(rtm.NumForcedGC),
		"NumGC":         int64(rtm.NumGC),
		"OtherSys":      int64(rtm.OtherSys),
		"PauseTotalNs":  int64(rtm.PauseTotalNs),
		"StackInuse":    int64(rtm.StackInuse),
		"StackSys":      int64(rtm.StackSys),
		"Sys":           int64(rtm.Sys),
		"TotalAlloc":    int64(rtm.TotalAlloc),
		"PollCount":     c.PollCount,
		"RandomValue":   int64(c.RandomValue * randomFactor),
	}
}

func (c *Collector) GetGaugeMetrics() map[string]float64 {
	return c.GaugeMetrics
}

func (c *Collector) GetCounterMetrics() map[string]int64 {
	return c.CounterMetrics
}
