package logger

import (
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(constants.Logger, log)
		c.Next()
	}
}

func GetLogger(c *gin.Context) *zap.Logger {
	if v, ok := c.Get(constants.Logger); ok {
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
