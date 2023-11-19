package root

import (
	"fmt"
	"strings"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/agent/collector"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/agent/uploader"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/logger"
	"github.com/pkg/errors"
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
		logger.Init()

		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return errors.Wrap(err, "can't get addr flag")
		}
		reportInterval, err := cmd.Flags().GetInt("reportInterval")
		if err != nil {
			return errors.Wrap(err, "can't get reportInterval flag")
		}
		pollInterval, err := cmd.Flags().GetInt("pollInterval")
		if err != nil {
			return errors.Wrap(err, "can't get pollInterval flag")
		}

		parts := strings.Split(addr, ":")
		if len(parts) < 2 || parts[1] == "" {
			return fmt.Errorf("you must provide a non-empty port number")
		}

		errorChan := make(chan error)
		c := collector.NewCollector(time.Duration(pollInterval)*time.Second, errorChan)
		u := uploader.NewUploader(addr, time.Duration(reportInterval)*time.Second,
			c.GetGaugeMetrics, c.GetCounterMetrics, errorChan)

		go c.Run()
		u.Run()

		return nil
	},
}
