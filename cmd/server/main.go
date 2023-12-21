package main

import (
	"fmt"

	"github.com/ElizavetaFirst/go-metrics-alerts/cmd/server/root"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/env"
)

const defaultStoringMetrics = 300

func main() {
	var addr string
	var storeInterval int
	var fileStoragePath string
	var restore bool
	var databaseDSN string
	root.RootCmd.PersistentFlags().StringVarP(&addr, "addr", "a",
		env.GetEnvString("ADDRESS", "localhost:8080"), "the address of the endpoint")
	root.RootCmd.PersistentFlags().IntVarP(&storeInterval, "storeInterval", "i",
		env.GetEnvDuration("STORE_INTERVAL", defaultStoringMetrics), "the frequency of storing metrics")
	root.RootCmd.PersistentFlags().StringVarP(&fileStoragePath, "fileStoragePath", "f",
		env.GetEnvString("FILE_STORAGE_PATH", "/tmp/metrics-db.json"), "the file storage path for storing metrics")
	root.RootCmd.PersistentFlags().BoolVarP(&restore, "restore", "r",
		env.GetEnvBool("RESTORE", true), "the flag to decide restore metrics from disk")
	root.RootCmd.PersistentFlags().StringVarP(&databaseDSN, "databaseDSN", "d",
		env.GetEnvString("DATABASE_DSN", ""),
		"db address")

	if err := root.RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
