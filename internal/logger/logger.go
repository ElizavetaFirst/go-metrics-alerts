package logger

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init() {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.Lock(zapcore.AddSync(os.Stderr)),
		zap.AtomicLevel{},
	))
	defer func() {
		if err := logger.Sync(); err != nil {
			if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "inappropriate ioctl for device") {
				logger.Error("Can't sync zap logger", zap.Error(err))
			}
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
