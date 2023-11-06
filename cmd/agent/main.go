package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/agent/collector"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/agent/uploader"
	"github.com/spf13/cobra"
)

var (
	addr           string
	reportInterval time.Duration
	pollInterval   time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "This is my agent application",
	Long:  "This is my agent application and it's has some long description",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			fmt.Printf("Unknown flags: %s\n", args)
			return fmt.Errorf("unknown flags: %s", args)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		parts := strings.Split(addr, ":")
		if len(parts) < 2 || parts[1] == "" {
			fmt.Println("You must provide a non-empty port number.")
			return
		}

		c := collector.NewCollector(pollInterval)
		u := uploader.NewUploader(addr, reportInterval, c.GetGaugeMetrics, c.GetCounterMetrics)

		go c.Run()
		u.Run()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", "localhost:8080", "the address of the endpoint")
	rootCmd.PersistentFlags().DurationVarP(&reportInterval, "reportInterval", "r", 10*time.Second, "the frequency of sending metrics to the server")
	rootCmd.PersistentFlags().DurationVarP(&pollInterval, "pollInterval", "p", 2*time.Second, "the frequency of polling metrics from the runtime package")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
