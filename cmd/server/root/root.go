package root

import (
	"fmt"
	"strings"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/saver"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/webserver"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "app",
	Short: "This is my application",
	Long:  "This is my application and it's has some long description",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return fmt.Errorf("can't get addr flag %w", err)
		}
		parts := strings.Split(addr, ":")
		if len(parts) < 2 || parts[1] == "" {
			return fmt.Errorf("you must provide a non-empty port number")
		}
		storeInterval, err := cmd.Flags().GetInt("storeInterval")
		if err != nil {
			return fmt.Errorf("can't get storeInterval flag %w", err)
		}
		fileStoragePath, err := cmd.Flags().GetString("fileStoragePath")
		if err != nil {
			return fmt.Errorf("can't get fileStoragePath flag %w", err)
		}
		restore, err := cmd.Flags().GetBool("restore")
		if err != nil {
			return fmt.Errorf("can't get restore flag %w", err)
		}
		databaseDSN, err := cmd.Flags().GetString("databaseDSN")
		if err != nil {
			return fmt.Errorf("can't get databaseDSN %w", err)
		}

		var s storage.Storage
		if databaseDSN != "" {
			s, err = storage.NewPostgresStorage(cmd.Context(), databaseDSN)

			if err != nil {
				return fmt.Errorf("failed to create the postgres storage %w", err)
			}
			defer func() {
				if err := s.Close(); err != nil {
					fmt.Printf("failed to close the postgres storage %v", err)
				}
			}()
		} else {
			s = storage.NewMemStorage()
			saver := saver.NewSaver(storeInterval, fileStoragePath, restore, s)
			errChan := make(chan error)

			go func() {
				if err := saver.Run(cmd.Context()); err != nil {
					errChan <- err
				}
				close(errChan)
			}()
			err := <-errChan
			if err != nil {
				return fmt.Errorf("error while saver Run %w", err)
			}
		}
		server := webserver.NewWebserver(s)

		return fmt.Errorf("error while server Run %w", server.Run(addr))
	},
}
