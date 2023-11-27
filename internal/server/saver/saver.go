package saver

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/logger"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"go.uber.org/zap"
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

func (s *Saver) getAndSaveMetrics() error {
	metrics := s.storage.GetAll()
	if len(metrics) == 0 {
		return nil
	}

	if err := saveMetricsToFile(metrics, s.fileStoragePath); err != nil {
		logger.GetLogger().Warn("can't save metrics to file", zap.String("fileStoragePath", s.fileStoragePath))
		return err
	}
	return nil
}

func (s *Saver) Run() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range c {
			if err := s.getAndSaveMetrics(); err != nil {
				logger.GetLogger().Error(fmt.Sprintf("can't save metrics on interrupt signal: %v", err))
			}
			os.Exit(0)
		}
	}()

	if s.restore {
		metrics, err := loadMetricsFromFile(s.fileStoragePath)
		if err != nil {
			return fmt.Errorf("cannot load metrics from file: %w", err)
		}
		fmt.Println("load", metrics)
		s.storage.SetAll(metrics)
	}

	ticker := time.NewTicker(s.storeInterval)

	errorCount := 0
	for range ticker.C {
		err := s.getAndSaveMetrics()
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("can't save metrics on timer tick: %v", err))
			errorCount++
		}
		if errorCount > constants.MaxErrors {
			return errors.New("too many errors in Saver:Run")
		}
	}

	if err := s.getAndSaveMetrics(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("can't save metrics when closing Saver: %v", err))
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
			logger.GetLogger().Error(fmt.Sprintf(failedCloseFile, closeErr))
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
			logger.GetLogger().Error(fmt.Sprintf(failedCloseFile, closeErr))
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}
	return metrics, nil
}
