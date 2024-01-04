package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/logger"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/metrics"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/gin-gonic/gin"
)

const (
	updateURL     = "/update/:metricType/:metricName/:metricValue"
	metricTypeStr = "metricType"
	metricNameStr = "metricName"
)

type Handler struct {
	Storage storage.Storage
}

func NewHandler(s storage.Storage) *Handler {
	return &Handler{
		Storage: s,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST(updateURL, logger.LogRequest(), h.handleUpdate)
	r.POST("/update", logger.LogRequest(), h.handleJSONUpdate)
	r.POST("/value/", logger.LogRequest(), h.handleJSONGetValue)
	r.POST("/updates/", logger.LogRequest(), h.handleUpdates)
	r.GET(updateURL, h.handleNotAllowed)
	r.GET("/value/:metricType/:metricName", logger.LogResponse(), h.handleGetValue)
	r.GET("/", logger.LogResponse(), h.handleGetAllValues)
	r.GET("/ping", logger.LogResponse(), h.handlePing)
}

func (h *Handler) handleJSONUpdate(c *gin.Context) {
	var metricType string
	var metricName string
	var metricValue any
	var metrics metrics.Metrics
	if err := c.ShouldBindJSON(&metrics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request: malformed JSON"})
		return
	}

	metricName = metrics.ID
	metricType = metrics.MType

	switch metricType {
	case constants.Gauge:
		metricValue = *metrics.Value
	case constants.Counter:
		metricValue = *metrics.Delta
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request: metricType should be gauge or counter"})
		return
	}

	if err := h.Storage.Update(c, &storage.UpdateOptions{
		MetricName: metricName,
		Update: storage.Metric{
			Type:  storage.MetricType(metricType),
			Value: metricValue,
		},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating metric"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) handleUpdate(c *gin.Context) {
	var metricType string
	var metricName string
	var metricValue any

	metricType = c.Param(metricTypeStr)
	metricName = c.Param(metricNameStr)
	metricValueParam := c.Param("metricValue")
	fmt.Println(metricType, metricName, metricValueParam)

	var err error
	switch metricType {
	case constants.Gauge:
		metricValue, err = strconv.ParseFloat(metricValueParam, 64)
	case constants.Counter:
		metricValue, err = strconv.ParseInt(metricValueParam, 10, 64)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	if err := h.Storage.Update(c, &storage.UpdateOptions{
		MetricName: metricName,
		Update: storage.Metric{
			Type:  storage.MetricType(metricType),
			Value: metricValue,
		},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating metric"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) handleNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method Not Allowed"})
}

func (h *Handler) handleJSONGetValue(c *gin.Context) {
	var metrics metrics.Metrics
	if err := c.ShouldBindJSON(&metrics); err != nil {
		c.Request.Context().Value(constants.LoggerKey{}).(*zap.Logger).Error("ShouldBindJSON return error",
			zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	value, err := h.Storage.Get(c, &storage.GetOptions{
		MetricName: metrics.ID,
		MetricType: metrics.MType,
	})

	if err != nil {
		if errors.Is(err, storage.ErrMetricNotFound) || string(value.Type) != metrics.MType {
			c.Status(http.StatusNotFound)
			return
		}
		c.Request.Context().Value(constants.LoggerKey{}).(*zap.Logger).Error("Get metric return error",
			zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	if metrics.MType == constants.Counter {
		delta, ok := value.Value.(int64)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data type for Delta"})
			return
		}
		metrics.Delta = &delta
	} else if metrics.MType == constants.Gauge {
		val, ok := value.Value.(float64)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data type for Value"})
			return
		}
		metrics.Value = &val
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) handleUpdates(c *gin.Context) {
	var metrics []metrics.Metrics

	if err := c.BindJSON(&metrics); err != nil {
		c.Request.Context().Value(constants.LoggerKey{}).(*zap.Logger).Error("BindJSON return error",
			zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	metricsMap := make(map[string]storage.Metric)
	for _, m := range metrics {
		switch m.MType {
		case constants.Gauge:
			metricsMap[m.ID] = storage.Metric{
				Value: m.Value,
				Type:  storage.MetricType(m.MType),
			}

		case constants.Counter:
			if m.Delta != nil {
				existingMetric, ok := metricsMap[m.ID]
				if ok {
					existingDelta, ok := existingMetric.Value.(*int64)
					if !ok {
						c.Request.Context().Value(constants.LoggerKey{}).(*zap.Logger).Error("all counter values must be int64")
						continue
					}
					newDelta := *existingDelta + *m.Delta
					existingMetric.Value = &newDelta
					metricsMap[m.ID] = existingMetric
				} else {
					delta := *m.Delta
					metricsMap[m.ID] = storage.Metric{
						Value: &delta,
						Type:  storage.MetricType(m.MType),
					}
				}
			}

		default:
			c.Request.Context().Value(constants.LoggerKey{}).(*zap.Logger).Error(
				"metrics can be only counter or gauge type, but this metric has incorrect type",
				zap.String("MetricType", m.MType))
		}
	}

	setAllOpts := storage.SetAllOptions{Metrics: metricsMap}
	err := h.Storage.SetAll(c.Request.Context(), &setAllOpts)

	if err != nil {
		c.Request.Context().Value(constants.LoggerKey{}).(*zap.Logger).Error("SetAll return error",
			zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) handleGetValue(c *gin.Context) {
	metricType := c.Param(metricTypeStr)
	metricName := c.Param(metricNameStr)

	value, err := h.Storage.Get(c, &storage.GetOptions{
		MetricName: metricName,
		MetricType: metricType,
	})
	if err != nil {
		if errors.Is(err, storage.ErrMetricNotFound) || string(value.Type) != metricType {
			c.Status(http.StatusNotFound)
			return
		}
		c.Request.Context().Value(constants.LoggerKey{}).(*zap.Logger).Error("Get metric return error",
			zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusOK, "%v", value.Value)
}

func (h *Handler) handleGetAllValues(c *gin.Context) {
	values, err := h.Storage.GetAll(c)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var htmlResponse strings.Builder

	htmlResponse.WriteString("<html><body>")
	for name, metric := range values {
		metricTypeStr := string(metric.Type)
		metricName := strings.TrimSuffix(name, metricTypeStr)
		htmlResponse.WriteString(fmt.Sprintf("<p>%s (%s): %v</p>", metricName, metric.Type, metric.Value))
	}
	htmlResponse.WriteString("</body></html>")

	c.Data(http.StatusOK, "text/html", []byte(htmlResponse.String()))
}

func (h *Handler) handlePing(c *gin.Context) {
	err := h.Storage.Ping(c)
	if err != nil {
		c.Request.Context().Value(constants.LoggerKey{}).(*zap.Logger).Error("Ping return error",
			zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
