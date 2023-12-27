package webserver

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/logger"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/middleware"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/handler"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/cenkalti/backoff"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	initialInterval = 1 * time.Second
	multiplier      = 2
	maxInterval     = 5 * time.Second
	maxElapsedTime  = 9 * time.Second
)

type Webserver struct {
	Router *gin.Engine
}

func NewWebserver(
	storage storage.Storage,
	log *zap.Logger,
) *Webserver {
	router := setupRouter(storage, log)

	return &Webserver{
		Router: router,
	}
}

func (ws *Webserver) Run(addr string) error {
	var err error

	operation := func() error {
		err = ws.Router.Run(addr)

		if err != nil {
			if errors.Is(err, storage.ErrDBNotInited) ||
				errors.Is(err, storage.ErrCantConnectDB) {
				return fmt.Errorf("server Run return retriable error %w", err)
			}

			return backoff.Permanent(err)
		}

		return nil
	}

	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.InitialInterval = initialInterval
	exponentialBackOff.Multiplier = multiplier
	exponentialBackOff.MaxInterval = maxInterval
	exponentialBackOff.MaxElapsedTime = maxElapsedTime

	if err := backoff.Retry(operation, exponentialBackOff); err != nil {
		return fmt.Errorf("retry failed %w", err)
	}

	return nil
}

func setupRouter(storage storage.Storage, log *zap.Logger) *gin.Engine {
	handler := handler.NewHandler(storage)

	r := gin.Default()
	r.Use(logger.InitLogger(log))
	r.Use(middleware.GzipGinRequestMiddleware)
	r.Use(func(c *gin.Context) {
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if strings.Contains(c.ContentType(), "application/json") ||
			strings.Contains(c.ContentType(), "text/html") ||
			strings.Contains(acceptEncoding, constants.Gzip) {
			gzip.Gzip(gzip.DefaultCompression)(c)
		}
	})

	handler.RegisterRoutes(r)

	return r
}
