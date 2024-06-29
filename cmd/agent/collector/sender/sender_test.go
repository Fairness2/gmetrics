package sender

import (
	"github.com/stretchr/testify/assert"
	"gmetrics/cmd/agent/env"
	"gmetrics/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	c := New()
	assert.NotNil(t, c)
}

func TestSendMetric(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() *httptest.Server
		metricType    string
		metricName    string
		metricValue   string
		expectedError bool
	}{
		{
			name: "successful metric send",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusOK)
				}))
			},
			metricType:    metrics.TypeGauge,
			metricName:    "TestMetric",
			metricValue:   "10",
			expectedError: false,
		},
		{
			name: "invalid http status",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusBadRequest)
				}))
			},
			metricType:    metrics.TypeGauge,
			metricName:    "TestMetric",
			metricValue:   "10",
			expectedError: true,
		},
		{
			name: "invalid http status",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusNotFound)
				}))
			},
			metricType:    metrics.TypeGauge,
			metricName:    "TestMetric",
			metricValue:   "10",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := tc.setupMock()
			//mockServer.URL = "http://127.0.0.1:8566"
			//mockServer.Start()
			env.ServerUrl = mockServer.URL

			defer mockServer.Close()

			c := New()
			err := c.sendMetric(tc.metricType, tc.metricName, tc.metricValue)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
