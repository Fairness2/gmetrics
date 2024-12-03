package sender

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/config"
	"gmetrics/cmd/agent/sendpool"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func getMockCollection() *collection.Type {
	cl := collection.NewCollection()
	cl.Values = map[string]any{
		"Alloc":         metrics.Gauge(1),
		"TotalAlloc":    metrics.Gauge(2),
		"BuckHashSys":   metrics.Gauge(3),
		"Frees":         metrics.Gauge(4),
		"GCCPUFraction": metrics.Gauge(5),
		"GCSys":         metrics.Gauge(6),
		"HeapAlloc":     metrics.Gauge(7),
		"HeapIdle":      metrics.Gauge(8),
		"HeapInuse":     metrics.Gauge(9),
		"HeapObjects":   metrics.Gauge(10),
		"HeapReleased":  metrics.Gauge(11),
		"HeapSys":       metrics.Gauge(12),
		"LastGC":        metrics.Gauge(13),
		"Lookups":       metrics.Gauge(14),
		"MCacheInuse":   metrics.Gauge(15),
		"MCacheSys":     metrics.Gauge(16),
		"MSpanInuse":    metrics.Gauge(17),
		"MSpanSys":      metrics.Gauge(18),
		"Mallocs":       metrics.Gauge(19),
		"NextGC":        metrics.Gauge(20),
		"NumForcedGC":   metrics.Gauge(21),
		"NumGC":         metrics.Gauge(22),
		"OtherSys":      metrics.Gauge(23),
		"PauseTotalNs":  metrics.Gauge(24),
		"StackInuse":    metrics.Gauge(25),
		"StackSys":      metrics.Gauge(26),
		"Sys":           metrics.Gauge(27),
		"RandomValue":   metrics.Gauge(29),
		"RandomCounter": metrics.Counter(29),
	}
	cl.PollCount = 28
	return cl
}

func TestNew(t *testing.T) {
	c := New(getMockCollection(), createMockSender(t))
	assert.NotNil(t, c)
}

func TestSendToServer(t *testing.T) {
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

			defer mockServer.Close()
			ctx, cancelFunc := context.WithCancel(context.TODO())
			defer cancelFunc()

			// Создаём пул отправок на сервер
			sendPool, poolErr := sendpool.New(ctx, 1, "1", mockServer.URL, nil)
			if poolErr != nil {
				assert.NoError(t, poolErr)
				return
			}
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

func createMockSender(t *testing.T) *MockSender {
	ctrl := gomock.NewController(t)
	s := NewMockSender(ctrl)

	return s
}

func TestSendMetrics(t *testing.T) {
	tests := []struct {
		name                 string
		sendToServerResponse *resty.Response
		sendToServerError    error
		wantError            bool
	}{
		{
			name:                 "successful_send_metrics",
			sendToServerResponse: &resty.Response{RawResponse: &http.Response{StatusCode: http.StatusOK}},
			sendToServerError:    nil,
			wantError:            false,
		},
		{
			name:                 "send_to_server_error",
			sendToServerResponse: nil,
			sendToServerError:    fmt.Errorf("some error"),
			wantError:            true,
		},
		{
			name:                 "unsuccessful_send_metrics",
			sendToServerResponse: &resty.Response{RawResponse: &http.Response{StatusCode: http.StatusNotFound}},
			sendToServerError:    nil,
			wantError:            true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSender := createMockSender(t)

			mockSender.EXPECT().
				Send(gomock.Any()).
				Return(tc.sendToServerResponse, tc.sendToServerError).
				Times(1)
			cl := getMockCollection()
			client := New(cl, mockSender)
			err := client.sendMetrics()

			if tc.wantError {
				assert.Error(t, err)
				assert.Equal(t, cl.PollCount, metrics.Counter(28))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, cl.PollCount, metrics.Counter(0))
			}
		})
	}
}

func TestRetrySend(t *testing.T) {
	tests := []struct {
		name       string
		wantError  bool
		getSenders func(t *testing.T) Sender
	}{
		{
			name:      "successful_send_metrics",
			wantError: false,
			getSenders: func(t *testing.T) Sender {
				mockSender := createMockSender(t)
				mockSender.EXPECT().
					Send(gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: http.StatusOK}}, nil).
					AnyTimes()
				return mockSender
			},
		},
		{
			name:      "send_to_server_not_repitable_error",
			wantError: true,
			getSenders: func(t *testing.T) Sender {
				mockSender := createMockSender(t)
				mockSender.EXPECT().
					Send(gomock.Any()).
					Return(nil, fmt.Errorf("some error")).
					AnyTimes()
				return mockSender
			},
		},
		{
			name:      "unsuccessful_send_metrics_repitable_error",
			wantError: true,
			getSenders: func(t *testing.T) Sender {
				mockSender := createMockSender(t)
				mockSender.EXPECT().
					Send(gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: http.StatusInternalServerError}}, nil).
					AnyTimes()
				return mockSender
			},
		},
		{
			name:      "unsuccessful_send_metrics_repitable_error_ones",
			wantError: false,
			getSenders: func(t *testing.T) Sender {
				mockSender := createMockSender(t)
				first := mockSender.EXPECT().
					Send(gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: http.StatusInternalServerError}}, nil).
					Times(1)
				mockSender.EXPECT().
					Send(gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: http.StatusOK}}, nil).
					After(first)
				return mockSender
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cl := getMockCollection()
			client := New(cl, tc.getSenders(t))
			client.retrySend()
		})
	}
}

func TestPeriodicSender(t *testing.T) {
	tests := []struct {
		name      string
		doneAfter time.Duration
	}{
		{
			name:      "sender_called_before_context_done",
			doneAfter: 3 * time.Second,
		},
		{
			name:      "sender_not_called_if_context_done_immediately",
			doneAfter: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			config.Params = &config.CliConfig{ReportInterval: 1}
			mockSender := NewMockSender(ctrl)
			mockSender.EXPECT().
				Send(gomock.Any()).
				Return(&resty.Response{RawResponse: &http.Response{StatusCode: http.StatusOK}}, nil).
				AnyTimes()
			cl := getMockCollection()
			client := New(cl, mockSender)

			ctx, cancel := context.WithTimeout(context.Background(), tc.doneAfter)
			defer cancel()
			client.PeriodicSender(ctx)
		})
	}
}
