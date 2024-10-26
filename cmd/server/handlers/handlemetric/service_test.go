package handlemetric

import (
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricByStringValue(t *testing.T) {
	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue string
		expectError bool
	}{
		{
			name:        "gauge_valid",
			metricType:  metrics.TypeGauge,
			metricName:  "Load",
			metricValue: "1.23",
			expectError: false,
		},
		{
			name:        "gauge_not_valid",
			metricType:  metrics.TypeGauge,
			metricName:  "Load",
			metricValue: "abc",
			expectError: true,
		},
		{
			name:        "counter_valid",
			metricType:  metrics.TypeCounter,
			metricName:  "Requests",
			metricValue: "10",
			expectError: false,
		},
		{
			name:        "counter_not_valid",
			metricType:  metrics.TypeCounter,
			metricName:  "Requests",
			metricValue: "string_number",
			expectError: true,
		},
		{
			name:        "type_not_valid",
			metricType:  "FileType",
			metricName:  "Docs",
			metricValue: "1",
			expectError: true,
		},
	}

	metrics.MeStore = metrics.NewMemStorage()
	// Run the tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := updateMetricByStringValue(tc.metricType, tc.metricName, tc.metricValue)
			if tc.expectError {
				assert.Error(t, err, "expect error")
			} else {
				assert.NoError(t, err, "expect no error")
			}
		})
	}
}

func TestUpdateMetricByRequestBody(t *testing.T) {
	tests := []struct {
		name        string
		payload     func() payload.Metrics
		expectError bool
	}{
		{
			name: "gauge_valid",
			payload: func() payload.Metrics {
				val := 1.23
				return payload.Metrics{
					ID:    "Load",
					MType: metrics.TypeGauge,
					Value: &val,
				}
			},
			expectError: false,
		},
		{
			name: "gauge_not_valid",
			payload: func() payload.Metrics {
				return payload.Metrics{
					ID:    "Load",
					MType: metrics.TypeGauge,
					Value: nil,
				}
			},
			expectError: true,
		},
		{
			name: "empty_id",
			payload: func() payload.Metrics {
				val := 1.23
				return payload.Metrics{
					ID:    "",
					MType: metrics.TypeGauge,
					Value: &val,
				}
			},
			expectError: true,
		},
		{
			name: "counter_valid",
			payload: func() payload.Metrics {
				var val int64 = 64
				return payload.Metrics{
					ID:    "Load",
					MType: metrics.TypeCounter,
					Delta: &val,
				}
			},
			expectError: false,
		},
		{
			name: "counter_not_valid",
			payload: func() payload.Metrics {
				return payload.Metrics{
					ID:    "Load",
					MType: metrics.TypeCounter,
					Delta: nil,
				}
			},
			expectError: true,
		},
		{
			name: "type_not_valid",
			payload: func() payload.Metrics {
				var val int64 = 64
				return payload.Metrics{
					ID:    "Load",
					MType: "FileType",
					Delta: &val,
				}
			},
			expectError: true,
		},
	}

	metrics.MeStore = metrics.NewMemStorage()
	// Run the tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := updateMetricByRequestBody(tc.payload())
			if tc.expectError {
				assert.Error(t, err, "expect error")
			} else {
				assert.NoError(t, err, "expect no error")
			}
		})
	}
}

func TestUpdateMetricsByRequestBody(t *testing.T) {
	tests := []struct {
		name        string
		payload     func() []payload.Metrics
		expectError bool
	}{
		{
			name: "gauge_valid",
			payload: func() []payload.Metrics {
				val := 1.23
				return []payload.Metrics{{
					ID:    "Load",
					MType: metrics.TypeGauge,
					Value: &val,
				}}
			},
			expectError: false,
		},
		{
			name: "gauge_not_valid",
			payload: func() []payload.Metrics {
				return []payload.Metrics{{
					ID:    "Load",
					MType: metrics.TypeGauge,
					Value: nil,
				}}
			},
			expectError: true,
		},
		{
			name: "empty_id",
			payload: func() []payload.Metrics {
				val := 1.23
				return []payload.Metrics{{
					ID:    "",
					MType: metrics.TypeGauge,
					Value: &val,
				}}
			},
			expectError: true,
		},
		{
			name: "counter_valid",
			payload: func() []payload.Metrics {
				var val int64 = 64
				return []payload.Metrics{{
					ID:    "Load",
					MType: metrics.TypeCounter,
					Delta: &val,
				}}
			},
			expectError: false,
		},
		{
			name: "counter_not_valid",
			payload: func() []payload.Metrics {
				return []payload.Metrics{{
					ID:    "Load",
					MType: metrics.TypeCounter,
					Delta: nil,
				}}
			},
			expectError: true,
		},
		{
			name: "type_not_valid",
			payload: func() []payload.Metrics {
				var val int64 = 64
				return []payload.Metrics{{
					ID:    "Load",
					MType: "FileType",
					Delta: &val,
				}}
			},
			expectError: true,
		},
		{
			name: "empty_body",
			payload: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			expectError: false,
		},
	}

	metrics.MeStore = metrics.NewMemStorage()
	// Run the tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := updateMetricsByRequestBody(tc.payload())
			if tc.expectError {
				assert.Error(t, err, "expect error")
			} else {
				assert.NoError(t, err, "expect no error")
			}
		})
	}
}
