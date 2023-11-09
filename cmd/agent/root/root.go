package root

import (
	"fmt"
	"strings"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/agent/collector"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/agent/uploader"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
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
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
		  return err
		}
		reportInterval, err := cmd.Flags().GetDuration("reportInterval")
		if err != nil {
		  return err
		}
		pollInterval, err := cmd.Flags().GetDuration("pollInterval")
		if err != nil {
		  return err
		}
	  
		parts := strings.Split(addr, ":")
		if len(parts) < 2 || parts[1] == "" {
			return fmt.Errorf("you must provide a non-empty port number")
		}

		c := collector.NewCollector(pollInterval)
		u := uploader.NewUploader(addr, reportInterval, c.GetGaugeMetrics, c.GetCounterMetrics)

		go c.Run()
		u.Run()

		return nil
	},
}
