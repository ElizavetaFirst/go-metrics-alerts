package root

import (
	"fmt"
	"strings"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/compressor"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/logger"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/handler"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "app",
	Short: "This is my application",
	Long:  "This is my application and it's has some long description",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			fmt.Printf("Unknown flags: %s\n", args)
			return fmt.Errorf("unknown flags: %s", args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Init()
		defer logger.Sync()

		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return errors.Wrap(err, "can't get addr flag")
		}
		parts := strings.Split(addr, ":")
		if len(parts) < 2 || parts[1] == "" {
			return fmt.Errorf("you must provide a non-empty port number")
		}

		r := gin.Default()
		r.Use(compressor.GzipGinRequestMiddleware)

		r.Use(func(c *gin.Context) {
			acceptEncoding := c.GetHeader("Accept-Encoding")
			if strings.Contains(c.ContentType(), "application/json") ||
				strings.Contains(c.ContentType(), "text/html") ||
				strings.Contains(acceptEncoding, constants.Gzip) {
				gzip.Gzip(gzip.DefaultCompression)(c)
			}
		})

		storage := storage.NewMemStorage()

		handler := handler.NewHandler(storage)

		handler.RegisterRoutes(r)

		err = r.Run(addr)
		if err != nil {
			return fmt.Errorf("run addr %s error %w", addr, err)
		}

		return nil
	},
}
