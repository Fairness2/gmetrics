// File: rpcmanyhandler_test.go
package handlemetric

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/metrics"
	pb "gmetrics/internal/payload/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestNewRPCManyHandler(t *testing.T) {
	got := NewRPCManyHandler()
	assert.NotNil(t, got)
}

func TestRPCManyHandler_HandleMetrics(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		wantStatus      codes.Code
		wantContentType string
	}{
		{
			name:            "empty_type",
			body:            `[{"id":"someName","type":"","value":123}]`,
			wantStatus:      codes.InvalidArgument,
			wantContentType: "application/json",
		},
		{
			name:            "empty_name",
			body:            `[{"id":"","type":"gauge","value":123}]`,
			wantStatus:      codes.InvalidArgument,
			wantContentType: "application/json",
		},
		{
			name:            "empty_value",
			body:            `[{"id":"someName","type":"gauge"}]`,
			wantStatus:      codes.InvalidArgument,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_type",
			body:            `[{"id":"someName","type":"aboba","value":123}]`,
			wantStatus:      codes.InvalidArgument,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_value_gauge",
			body:            `[{"id":"someName","type":"gauge","value":"some"}]`,
			wantStatus:      codes.InvalidArgument,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_value_count",
			body:            `[{"id":"someName","type":"counter","delta":"some"}]`,
			wantStatus:      codes.InvalidArgument,
			wantContentType: "application/json",
		},
		{
			name:            "right_value_gauge",
			body:            `[{"id":"someName","type":"gauge","value":56.78}]`,
			wantStatus:      codes.OK,
			wantContentType: "application/json",
		},
		{
			name:            "right_value_count",
			body:            `[{"id":"someName","type":"counter","delta":5}]`,
			wantStatus:      codes.OK,
			wantContentType: "application/json",
		},
	}

	// Устанавливаем глобальное хранилище метрик
	storage := metrics.NewMemStorage()
	metrics.MeStore = storage

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			service := NewRPCManyHandler()
			req := &pb.MetricsRequest{Body: []byte(test.body)}
			_, err := service.HandleMetrics(context.TODO(), req)
			code := status.Code(err)
			assert.Equal(t, test.wantStatus, code, "unexpected error code")
		})
	}
}
