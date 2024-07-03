package sender

import (
	"github.com/stretchr/testify/assert"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/env"
	"gmetrics/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getMockCollection() *collection.CollectionType {
	return &collection.CollectionType{
		Alloc:         1,
		TotalAlloc:    2,
		BuckHashSys:   3,
		Frees:         4,
		GCCPUFraction: 5,
		GCSys:         6,
		HeapAlloc:     7,
		HeapIdle:      8,
		HeapInuse:     9,
		HeapObjects:   10,
		HeapReleased:  11,
		HeapSys:       12,
		LastGC:        13,
		Lookups:       14,
		MCacheInuse:   15,
		MCacheSys:     16,
		MSpanInuse:    17,
		MSpanSys:      18,
		Mallocs:       19,
		NextGC:        20,
		NumForcedGC:   21,
		NumGC:         22,
		OtherSys:      23,
		PauseTotalNs:  24,
		StackInuse:    25,
		StackSys:      26,
		Sys:           27,
		PollCount:     28,
		RandomValue:   29,
	}
}

func TestNew(t *testing.T) {
	c := New(getMockCollection())
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
			env.ServerURL = mockServer.URL

			defer mockServer.Close()

			c := New(getMockCollection())
			err := c.sendMetric(tc.metricType, tc.metricName, tc.metricValue)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
