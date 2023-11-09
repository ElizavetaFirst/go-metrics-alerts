package main

import (
	"fmt"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/cmd/agent/root"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/env"
)

var (
	addr           string
	reportInterval time.Duration
	pollInterval   time.Duration
)

func init() {
	addr = env.GetEnvString("ADDRESS", "localhost:8080")
	reportInterval = env.GetEnvDuration("REPORT_INTERVAL", 10*time.Second)
	pollInterval = env.GetEnvDuration("POLL_INTERVAL", 2*time.Second)

	root.RootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", addr, "the address of the endpoint")
	root.RootCmd.PersistentFlags().DurationVarP(&reportInterval, "reportInterval", "r", reportInterval, "the frequency of sending metrics to the server")
	root.RootCmd.PersistentFlags().DurationVarP(&pollInterval, "pollInterval", "p", pollInterval, "the frequency of polling metrics from the runtime package")
}

func main() {
	if err := root.RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
