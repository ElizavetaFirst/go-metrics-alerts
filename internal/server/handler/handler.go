package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Storage storage.Storage
}

func NewHandler(s storage.Storage) *Handler {
	return &Handler{Storage: s}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/update/:metricType/:metricName/:metricValue", h.handleUpdate)
	r.GET("/update/:metricType/:metricName/:metricValue", h.handleNotAllowed)
	r.GET("/value/:metricType/:metricName", h.handleGetValue)
	r.GET("/", h.handleGetAllValues)

}

func (h *Handler) handleUpdate(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")
	metricValueParam := c.Param("metricValue")

	var metricValue any
	var err error
	switch metricType {
	case "gauge":
		metricValue, err = strconv.ParseFloat(metricValueParam, 64)
	case "counter":
		metricValue, err = strconv.ParseInt(metricValueParam, 10, 64)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	if err := h.Storage.Update(metricName, storage.Metric{Type: storage.MetricType(metricType), Value: metricValue}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating metric"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) handleNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method Not Allowed"})
}

// обработчик для GET запроса на получение значения метрики
func (h *Handler) handleGetValue(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")

	value, found := h.Storage.Get(metricName)
	if !found || string(value.Type) != metricType {
		c.Status(http.StatusNotFound)
		return
	}

	c.String(http.StatusOK, "%v", value.Value) // возвращаем значение
}

func (h *Handler) handleGetAllValues(c *gin.Context) {
	values := h.Storage.GetAll()

	// создаем HTML ответ
	htmlResponse := "<html><body>"
	for name, metric := range values {
		htmlResponse += fmt.Sprintf("<p>%s (%s): %v</p>", name, metric.Type, metric.Value)
	}
	htmlResponse += "</body></html>"

	c.Data(http.StatusOK, "text/html", []byte(htmlResponse))
}
