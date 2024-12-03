package handlemetric

import (
	"context"
	"encoding/json"
	"errors"
	"gmetrics/internal/logger"
	"gmetrics/internal/payload"
	pb "gmetrics/internal/payload/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RPCManyHandler Сервис для обновления метрик по rpc
type RPCManyHandler struct {
	pb.UnimplementedMetricsServiceServer
}

// HandleMetrics обновление метрик
func (r *RPCManyHandler) HandleMetrics(ctx context.Context, request *pb.MetricsRequest) (*pb.MetricsResponse, error) {
	rawBody := request.GetBody()
	// Парсим тело в структуру запроса
	var body []payload.Metrics
	err := json.Unmarshal(rawBody, &body)
	if err != nil {
		logger.Log.Infow("Bad request for update metric", "error", err, "body", string(rawBody))
		return nil, status.Error(codes.InvalidArgument, BadRequestError.Error())
	}
	var metricErr *UpdateMetricError
	uError := updateMetricsByRequestBody(body)
	if uError != nil {
		if errors.As(uError, &metricErr) {
			return nil, status.Error(codes.InvalidArgument, metricErr.Error())
		} else {
			return nil, status.Error(codes.Internal, uError.Error())
		}
	}

	return &pb.MetricsResponse{
		Status:  payload.ResponseSuccessStatus,
		Message: "",
	}, nil
}

// NewRPCManyHandler создание нового сервиса
func NewRPCManyHandler() *RPCManyHandler {
	return &RPCManyHandler{}
}
