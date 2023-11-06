package main

import (
	"fmt"
	"strings"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/handler"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var addr string

var rootCmd = &cobra.Command{
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
	Run: func(cmd *cobra.Command, args []string) {
		parts := strings.Split(addr, ":")
		if len(parts) < 2 || parts[1] == "" {
			fmt.Println("You must provide a non-empty port number.")
			return
		}

		r := gin.Default()

		storage := storage.NewMemStorage()

		handler := handler.NewHandler(storage)

		handler.RegisterRoutes(r)

		err := r.Run(addr)
		if err != nil {
			fmt.Printf("run addr %s error %v", addr, err)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", "localhost:8080", "the address of the endpoint")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
