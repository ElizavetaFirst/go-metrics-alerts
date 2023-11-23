package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElizavetaFirst/go-metrics-alerts/internal/constants"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockStorage struct{}

func (ms *mockStorage) Update(name string, metric storage.Metric) error {
	return nil
}

func (ms *mockStorage) Get(name string, metricType string) (storage.Metric, bool) {
	return storage.Metric{}, false
}

func (ms *mockStorage) GetAll() map[string]storage.Metric {
	return map[string]storage.Metric{
		"test": {Type: constants.Gauge, Value: 123},
	}
}

func TestHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int // expected HTTP code
	}{
		{
			"Valid request",
			http.MethodPost,
			"/update/gauge/test/123",
			http.StatusOK,
		},
		{
			"Invalid method request",
			http.MethodGet,
			"/update/gauge/test/123",
			http.StatusMethodNotAllowed,
		},
		{
			"Not found",
			http.MethodPost,
			"/update/test/123",
			http.StatusNotFound,
		},
		{
			"Bad request",
			http.MethodPost,
			"/update/counter/test/123.3",
			http.StatusBadRequest,
		},
		{
			"Not found get value",
			http.MethodGet,
			"/value/gauge/nonexistent",
			http.StatusNotFound,
		},
		{
			"Valid get all values request",
			http.MethodGet,
			"/",
			http.StatusOK,
		},
	}

	ms := &mockStorage{}
	h := NewHandler(ms)

	r := gin.Default()
	h.RegisterRoutes(r)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = req

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
