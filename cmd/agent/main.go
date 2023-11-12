package main

import (
	"fmt"

	"github.com/ElizavetaFirst/go-metrics-alerts/cmd/agent/root"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/env"
)

var (
	addr           string
	reportInterval int
	pollInterval   int
)

func main() {
	addr = env.GetEnvString("ADDRESS", "localhost:8080")
	reportInterval = env.GetEnvDuration("REPORT_INTERVAL", 10)
	pollInterval = env.GetEnvDuration("POLL_INTERVAL", 2)

	root.RootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", addr, "the address of the endpoint")
	root.RootCmd.PersistentFlags().IntVarP(&reportInterval, "reportInterval", "r", reportInterval, "the frequency of sending metrics to the server")
	root.RootCmd.PersistentFlags().IntVarP(&pollInterval, "pollInterval", "p", pollInterval, "the frequency of polling metrics from the runtime package")

	if err := root.RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
