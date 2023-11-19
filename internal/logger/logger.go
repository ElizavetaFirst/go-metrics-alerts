package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var log *zap.Logger

func Init() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("can't initialize zap logger: %v", err))
	}
	defer func() {
		if err := log.Sync(); err != nil {
			panic(fmt.Sprintf("Can't sync zap logger: %v", err))
		}
	}()
}

func GetLogger() *zap.Logger {
	return log
}

// LogRequest is a middleware that logs HTTP requests.
func LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetLogger() == nil {
			return
		}
		start := time.Now()

		// Pass to the next middleware/handler
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		GetLogger().Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			zap.String("duration", duration.String()),
		)
	}
}

func LogResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetLogger() == nil {
			return
		}
		// Call the next middleware or handler
		c.Next()

		// Log the information about the response
		GetLogger().Info("HTTP Response",
			zap.Int("status", c.Writer.Status()),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}
