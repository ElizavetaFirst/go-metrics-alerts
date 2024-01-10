package main

import (
	"log"

	"github.com/ElizavetaFirst/go-metrics-alerts/cmd/agent/root"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/env"
)

const (
	defaultAddress        = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
)

func main() {
	addr := env.GetEnvString("ADDRESS", defaultAddress)
	reportInterval := env.GetEnvDuration("REPORT_INTERVAL", defaultReportInterval)
	pollInterval := env.GetEnvDuration("POLL_INTERVAL", defaultPollInterval)

	root.RootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", addr, "the address of the endpoint")
	root.RootCmd.PersistentFlags().IntVarP(&reportInterval, "reportInterval", "r", reportInterval,
		"the frequency of sending metrics to the server")
	root.RootCmd.PersistentFlags().IntVarP(&pollInterval, "pollInterval", "p", pollInterval,
		"the frequency of polling metrics from the runtime package")

	if err := root.RootCmd.Execute(); err != nil {
		log.Println(err)
	}
}
