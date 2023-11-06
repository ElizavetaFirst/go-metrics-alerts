package main

import (
	"fmt"
	"os"
	"strconv"
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

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if envVal, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(envVal); err == nil {
			return time.Duration(i) * time.Second
		}
	}
	return defaultVal
}

func getEnvString(key, defaultVal string) string {
	if envVal, exists := os.LookupEnv(key); exists {
		return envVal
	}
	return defaultVal
}

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
	addr = getEnvString("ADDRESS", "localhost:8080")
	reportInterval = getEnvDuration("REPORT_INTERVAL", 10*time.Second)
	pollInterval = getEnvDuration("POLL_INTERVAL", 2*time.Second)

	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", addr, "the address of the endpoint")
	rootCmd.PersistentFlags().DurationVarP(&reportInterval, "reportInterval", "r", reportInterval, "the frequency of sending metrics to the server")
	rootCmd.PersistentFlags().DurationVarP(&pollInterval, "pollInterval", "p", pollInterval, "the frequency of polling metrics from the runtime package")

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
