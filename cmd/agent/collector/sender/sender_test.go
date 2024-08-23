package sender

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/collector/sender/mock"
	"gmetrics/cmd/agent/config"
	"gmetrics/cmd/agent/sendpool"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getMockCollection() *collection.Type {
	return &collection.Type{
		Values: map[string]any{
			"Alloc":         1,
			"TotalAlloc":    2,
			"BuckHashSys":   3,
			"Frees":         4,
			"GCCPUFraction": 5,
			"GCSys":         6,
			"HeapAlloc":     7,
			"HeapIdle":      8,
			"HeapInuse":     9,
			"HeapObjects":   10,
			"HeapReleased":  11,
			"HeapSys":       12,
			"LastGC":        13,
			"Lookups":       14,
			"MCacheInuse":   15,
			"MCacheSys":     16,
			"MSpanInuse":    17,
			"MSpanSys":      18,
			"Mallocs":       19,
			"NextGC":        20,
			"NumForcedGC":   21,
			"NumGC":         22,
			"OtherSys":      23,
			"PauseTotalNs":  24,
			"StackInuse":    25,
			"StackSys":      26,
			"Sys":           27,
			"RandomValue":   29,
		},
		PollCount: 28,
	}
}

func TestNew(t *testing.T) {
	c := New(getMockCollection(), createMockPusher(t))
	assert.NotNil(t, c)
}

func TestSendMetric(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() *httptest.Server
		body          func() []payload.Metrics
		metricType    string
		metricName    string
		metricValue   string
		expectedError bool
	}{
		{
			name: "successful_metric_send",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusOK)
				}))
			},
			body: func() []payload.Metrics {
				var val float64 = 10
				return []payload.Metrics{{
					ID:    "TestMetric",
					MType: metrics.TypeGauge,
					Delta: nil,
					Value: &val,
				}}
			},
			expectedError: false,
		},
		{
			name: "HTTP_status_bad_request",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusBadRequest)
				}))
			},
			body: func() []payload.Metrics {
				var val float64 = 10
				return []payload.Metrics{{
					ID:    "TestMetric",
					MType: metrics.TypeGauge,
					Delta: nil,
					Value: &val,
				}}
			},
			expectedError: true,
		},
		{
			name: "HTTP_status_not_found",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusNotFound)
				}))
			},
			body: func() []payload.Metrics {
				var val float64 = 10
				return []payload.Metrics{{
					ID:    "TestMetric",
					MType: metrics.TypeGauge,
					Delta: nil,
					Value: &val,
				}}
			},
			expectedError: true,
		},
	}
	config.Params = config.InitializeDefaultConfig()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := tc.setupMock()
			config.Params.ServerURL = mockServer.URL

			defer mockServer.Close()
			ctx, cancelFunc := context.WithCancel(context.TODO())
			defer cancelFunc()

			// Создаём пул отправок на сервер
			sendPool := sendpool.New(ctx, 1)
			c := New(getMockCollection(), sendPool)
			err := c.sendToServer(tc.body())
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createMockPusher(t *testing.T) *mock.MockPusher {
	ctrl := gomock.NewController(t)
	s := mock.NewMockPusher(ctrl)

	return s
}
