package saver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
)

const failedCloseFile = "Failed to close file: %v"

type Saver struct {
	log             *zap.Logger
	storage         storage.Storage
	fileStoragePath string
	storeInterval   time.Duration
	restore         bool
}

func NewSaver(storeInterval int,
	fileStoragePath string,
	restore bool,
	storage storage.Storage,
	log *zap.Logger,
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
		s.log.Warn("can't GetAll metrics", zap.Error(err))
	}
	if len(metrics) == 0 {
		return nil
	}

	if err := saveMetricsToFile(metrics, s.fileStoragePath); err != nil {
		s.log.Warn("can't save metrics to file",
			zap.String("fileStoragePath", s.fileStoragePath))
		return err
	}
	return nil
}

func (s *Saver) Run(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() {
		if err := s.getAndSaveMetrics(ctx); err != nil {
			s.log.Warn("can't save metrics on interrupt signal", zap.Error(err))
		}
	}()

	if s.restore {
		metrics, err := loadMetricsFromFile(s.fileStoragePath)
		if err != nil {
			s.log.Warn("cannot load metrics from file", zap.Error(err))
		}
		err = s.storage.SetAll(ctx, &storage.SetAllOptions{Metrics: metrics})
		if err != nil {
			s.log.Warn("cannot set all metrics", zap.Error(err))
			return fmt.Errorf("cannot set all metrics: %w", err)
		}
	}

	ticker := time.NewTicker(s.storeInterval)
	defer ticker.Stop()

	errorCount := 0
	for {
		select {
		case <-ticker.C:
			err := s.getAndSaveMetrics(ctx)
			if err != nil {
				s.log.Warn("can't save metrics on timer tick",
					zap.Error(err))
				errorCount++
			}
			if errorCount > constants.MaxErrors {
				return errors.New("too many errors in Saver:Run")
			}
		case <-ctx.Done():
			if err := s.getAndSaveMetrics(ctx); err != nil {
				s.log.Warn("can't save metrics when closing Saver",
					zap.Error(err))
			}
			return fmt.Errorf("saver run() context return error %w", ctx.Err())
		}
	}
}

func saveMetricsToFile(metrics map[string]storage.Metric, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			log.Printf(failedCloseFile, closeErr)
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
			log.Printf(failedCloseFile, closeErr)
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}
	return metrics, nil
}
