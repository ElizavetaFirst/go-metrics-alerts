package logger

import (
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const logger = "Logger"

func InitLogger() gin.HandlerFunc {
	var log *zap.Logger
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("can't initialize zap logger: %v", err))
	}

	return func(c *gin.Context) {
		c.Set(logger, log)
		c.Next()

		// Call Sync at the end of every request
		if err := log.Sync(); err != nil && (!errors.Is(err, syscall.EBADF) && !errors.Is(err, syscall.ENOTTY)) {
			panic(fmt.Sprintf("can't sync zap logger: %v", err))
		}
	}
}

func GetLogger(c *gin.Context) *zap.Logger {
	if v, ok := c.Get(logger); ok {
		if log, ok := v.(*zap.Logger); ok {
			return log
		}
	}
	return nil
}

func LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := GetLogger(c)
		if log == nil {
			return
		}
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		log.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			zap.String("duration", duration.String()),
		)
	}
}

func LogResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := GetLogger(c)
		if log == nil {
			return
		}
		c.Next()

		log.Info("HTTP Response",
			zap.Int("status", c.Writer.Status()),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}
