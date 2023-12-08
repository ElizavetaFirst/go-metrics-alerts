package webserver

import (
	"strings"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/logger"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/middleware"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/db"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/handler"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type Webserver struct {
	Router *gin.Engine
}

func NewWebserver(
	storage storage.Storage,
	DB *db.DB,
) *Webserver {
	router := setupRouter(storage, DB)

	return &Webserver{
		Router: router,
	}
}

func (ws *Webserver) Run(addr string) error {
	return errors.Wrap(ws.Router.Run(addr), "error while Webserver Run")
}

func setupRouter(storage storage.Storage, DB *db.DB) *gin.Engine {
	handler := handler.NewHandler(storage, DB)

	r := gin.Default()
	r.Use(logger.InitLogger())
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
