package main

import (
	"log"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/handler"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	storage := storage.NewMemStorage()

	handler := handler.NewHandler(storage)

	r := gin.Default()

	handler.RegisterRoutes(r)

	log.Fatal(r.Run(":8080"))
}
