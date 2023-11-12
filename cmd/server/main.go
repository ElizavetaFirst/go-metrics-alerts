package main

import (
	"fmt"

	"github.com/ElizavetaFirst/go-metrics-alerts/cmd/server/root"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/env"
)

var addr string

func main() {
	root.RootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", env.GetEnvString("ADDRESS", "localhost:8080"), "the address of the endpoint")

	if err := root.RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
