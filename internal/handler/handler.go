package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/storage"
)

type Handler struct {
	Storage storage.Storage
}

func NewHandler(s storage.Storage) *Handler {
	return &Handler{Storage: s}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm := time.Date(2023, time.February, 21, 2, 51, 35, 0, time.UTC)
	w.Header().Set("Date", tm.Format(http.TimeFormat))
	w.Header().Set("Content-Length", "0")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Убеждаемся, что метод запроса - POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 5 || parts[1] != "update" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	metricType := parts[2]
	if metricType != "counter" && metricType != "gauge" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	metricName := parts[3]
	if metricName == "" {
		http.Error(w, "Status not found", http.StatusNotFound)
	}

	var metricValue float64
	var err error
	if metricValue, err = strconv.ParseFloat(parts[4], 64); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := h.Storage.Update(metricName, storage.Metric{Type: storage.MetricType(metricType), Value: metricValue}); err != nil {
		http.Error(w, "Error updating metric", http.StatusInternalServerError)
		return
	}

	// Возвращает пользователя в ответе
	w.WriteHeader(http.StatusOK)
}
