package root

import (
	"fmt"
	"os"
	"strings"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/saver"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/webserver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "app",
	Short: "This is my application",
	Long:  "This is my application and it's has some long description",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return errors.Wrap(err, "can't get addr flag")
		}
		parts := strings.Split(addr, ":")
		if len(parts) < 2 || parts[1] == "" {
			return fmt.Errorf("you must provide a non-empty port number")
		}
		storeInterval, err := cmd.Flags().GetInt("storeInterval")
		if err != nil {
			return errors.Wrap(err, "can't get storeInterval flag")
		}
		fileStoragePath, err := cmd.Flags().GetString("fileStoragePath")
		if err != nil {
			return errors.Wrap(err, "can't get fileStoragePath flag")
		}
		restore, err := cmd.Flags().GetBool("restore")
		if err != nil {
			return errors.Wrap(err, "can't get restore flag")
		}
		databaseDSN, err := cmd.Flags().GetString("databaseDSN")
		if err != nil {
			return errors.Wrap(err, "can't get databaseDSN")
		}

		var s storage.Storage
		if databaseDSN != "" {
			s, err = storage.NewPostgresStorage(cmd.Context(), databaseDSN)
			
			if err != nil {
				return errors.Wrap(err, "failed to create the postgres storage")
			}
			defer func() {
				if err := s.Close(); err != nil {
					fmt.Printf("failed to close the postgres storage %v", err)
				}
			}()
		} else {
			s = storage.NewMemStorage()
		}

		saver := saver.NewSaver(storeInterval, fileStoragePath, restore, s)
		server := webserver.NewWebserver(s)
		go func() {
			if err := saver.Run(cmd.Context()); err != nil {
				fmt.Printf("error while saver Run %v", err)
				os.Exit(0)
			}
		}()

		return errors.Wrap(server.Run(addr), "error while server Run")
	},
}
