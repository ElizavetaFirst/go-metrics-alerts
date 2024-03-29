package root

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/saver"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/webserver"
	"github.com/spf13/cobra"
)

const timeoutShutdown = time.Second * 10

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

		ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancelCtx()

		context.AfterFunc(ctx, func() {
			ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
			defer cancelCtx()

			<-ctx.Done()
			log.Fatal("failed to gracefully shutdown the service")
		})

		wg := &sync.WaitGroup{}
		defer func() {
			wg.Wait()
		}()

		log, err := zap.NewProduction()
		if err != nil {
			fmt.Printf("can't initialize zap logger: %v", err)
			return fmt.Errorf("zap.NewProduction() return error %w", err)
		}
		defer func() {
			if err := log.Sync(); err != nil {
				log.Error("failed to Sync() log", zap.Error(err))
			}
		}()

		var s storage.Storage
		if databaseDSN != "" {
			s, err = storage.NewPostgresStorage(ctx, databaseDSN, log)

			if err != nil {
				return fmt.Errorf("failed to create the postgres storage %w", err)
			}
			defer func() {
				if err := s.Close(); err != nil {
					log.Error("failed to close the postgres storage", zap.Error(err))
				}
			}()
		} else {
			s = storage.NewMemStorage(log)
			saver := saver.NewSaver(storeInterval, fileStoragePath, restore, s, log)
			errChan := make(chan error)

			go func() {
				if err := saver.Run(ctx, wg); err != nil {
					errChan <- err
				}
				close(errChan)
			}()
			err := <-errChan
			if err != nil {
				return fmt.Errorf("error while saver Run %w", err)
			}
		}
		server := webserver.NewWebserver(s, log)

		return fmt.Errorf("error while server Run %w", server.Run(addr, wg))
	},
}
