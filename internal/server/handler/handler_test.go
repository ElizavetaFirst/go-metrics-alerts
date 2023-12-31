package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (ms *mockStorage) SetAll(metrics map[string]storage.Metric) {}

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

func TestSendCounterMetricsJson(t *testing.T) {
	mockServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Error("Expected POST request, got ", r.Method)
		}

		expectedURL := "/metrics"
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Wrong URL: got %v want %v", r.URL.EscapedPath(), expectedURL)
		}

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed reading request body: %v", err)
		}
		expectedBody := `{"test":123}`
		if string(reqBody) != expectedBody {
			t.Errorf("Unexpected body: got %v want %v", reqBody, expectedBody)
		}
	})

	req, err := http.NewRequest(http.MethodPost, "/metrics", strings.NewReader(`{"test":123}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := mockServer

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}
}
