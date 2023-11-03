package main

import (
	"log"
	"net/http"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/handler"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
)

func main() {
	storage := storage.NewMemStorage()

	handler := handler.NewHandler(storage)

	http.Handle("/update/", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
