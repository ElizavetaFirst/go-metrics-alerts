package saver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
)

const failedCloseFile = "Failed to close file: %v"

type Saver struct {
	storage         storage.Storage
	fileStoragePath string
	storeInterval   time.Duration
	restore         bool
}

func NewSaver(storeInterval int,
	fileStoragePath string,
	restore bool,
	storage storage.Storage,
) *Saver {
	return &Saver{
		storeInterval:   time.Duration(storeInterval) * time.Second,
		fileStoragePath: fileStoragePath,
		restore:         restore,
		storage:         storage,
	}
}

func (s *Saver) getAndSaveMetrics(ctx context.Context) error {
	metrics, err := s.storage.GetAll(ctx)
	if err != nil {
		ctx.Value(constants.Logger).(*zap.Logger).Warn("can't GetAll metrics", zap.Error(err))
	}
	if len(metrics) == 0 {
		return nil
	}

	if err := saveMetricsToFile(metrics, s.fileStoragePath); err != nil {
		ctx.Value(constants.Logger).(*zap.Logger).Warn("can't save metrics to file",
			zap.String("fileStoragePath", s.fileStoragePath))
		return err
	}
	return nil
}

func (s *Saver) Run(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range c {
			if err := s.getAndSaveMetrics(ctx); err != nil {
				ctx.Value(constants.Logger).(*zap.Logger).Warn("can't save metrics on interrupt signal", zap.Error(err))
			}
			os.Exit(0)
		}
	}()

	if s.restore {
		metrics, err := loadMetricsFromFile(s.fileStoragePath)
		if err != nil {
			ctx.Value(constants.Logger).(*zap.Logger).Warn("cannot load metrics from file", zap.Error(err))
		}
		err = s.storage.SetAll(ctx, &storage.SetAllOptions{Metrics: metrics})
		if err != nil {
			ctx.Value(constants.Logger).(*zap.Logger).Warn("cannot set all metrics", zap.Error(err))
			return fmt.Errorf("cannot set all metrics: %w", err)
		}
	}

	ticker := time.NewTicker(s.storeInterval)

	errorCount := 0
	for range ticker.C {
		err := s.getAndSaveMetrics(ctx)
		if err != nil {
			ctx.Value(constants.Logger).(*zap.Logger).Warn("can't save metrics on timer tick",
				zap.Error(err))
			errorCount++
		}
		if errorCount > constants.MaxErrors {
			return errors.New("too many errors in Saver:Run")
		}
	}

	if err := s.getAndSaveMetrics(ctx); err != nil {
		ctx.Value(constants.Logger).(*zap.Logger).Warn("can't save metrics when closing Saver",
			zap.Error(err))
	}
	return nil
}

func saveMetricsToFile(metrics map[string]storage.Metric, filePath string) error {
	fmt.Println(metrics)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fmt.Printf(failedCloseFile, closeErr)
		}
	}()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(metrics); err != nil {
		return fmt.Errorf("failed to encode metrics: %w", err)
	}

	return nil
}

func loadMetricsFromFile(filePath string) (map[string]storage.Metric, error) {
	metrics := make(map[string]storage.Metric)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return metrics, nil
		} else {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fmt.Printf(failedCloseFile, closeErr)
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}
	return metrics, nil
}
