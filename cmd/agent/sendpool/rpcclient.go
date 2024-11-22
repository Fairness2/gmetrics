package sendpool

import (
	"context"
	"errors"
	pb "gmetrics/internal/payload/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
	"strings"
)

// ErrorMethodNotExists Ошибка, что метод для отправки запроса по rpc ещё не реализован
var ErrorMethodNotExists = errors.New("method not exists")

// statusMap мапа статусов, чтобы сопоставлять ответы rpc с ответами по http
var statusMap = map[codes.Code]int{
	codes.OK:                 http.StatusOK,
	codes.Canceled:           http.StatusRequestTimeout,
	codes.Unknown:            http.StatusInternalServerError,
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.DeadlineExceeded:   http.StatusGatewayTimeout,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.ResourceExhausted:  http.StatusTooManyRequests,
	codes.FailedPrecondition: http.StatusPreconditionFailed,
	codes.Aborted:            http.StatusConflict,
	codes.OutOfRange:         http.StatusBadRequest,
	codes.Unimplemented:      http.StatusNotImplemented,
	codes.Internal:           http.StatusInternalServerError,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.DataLoss:           http.StatusUnsupportedMediaType,
	codes.Unauthenticated:    http.StatusUnauthorized,
}

// RPCResponse Специальный тип ответа для сопоставления с ответами по http
type RPCResponse struct {
	status int
}

// StatusCode возвращает статус код
func (R RPCResponse) StatusCode() int {
	return R.status
}

// NewRPCResponse создаём новый ответ с маппленным кодом, как в http
func NewRPCResponse(code codes.Code) RPCResponse {
	rCode, ok := statusMap[code]
	if !ok {
		rCode = http.StatusTeapot
	}
	return RPCResponse{
		status: rCode,
	}
}

// RPCConnection Интерфейс подключения по rpc
type RPCConnection interface {
	grpc.ClientConnInterface
	io.Closer
}

// RPCClient Клиент для общения с сервером по rpc
type RPCClient struct {
	ctx     context.Context
	conn    RPCConnection
	service pb.MetricsServiceClient
	netAddr string // реальный адрес кликета, будет встроен в X-Real-IP
}

// NewRPCClient Создание нового rpc клиента
func NewRPCClient(ctx context.Context, baseURL string) (*RPCClient, error) {
	baseURL = clearURL(baseURL)
	conn, err := grpc.NewClient(baseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	addr, err := getNetAddr()
	if err != nil {
		return nil, err
	}
	return &RPCClient{
		ctx:     ctx,
		conn:    conn,
		service: pb.NewMetricsServiceClient(conn),
		netAddr: addr,
	}, nil
}

// Post Отправка зпроса на сервер по rpc
func (r RPCClient) Post(url string, body []byte, headers ...Header) (MetricResponse, error) {
	requestCtx := r.createMeta(headers)
	var err error
	switch url {
	case URLUpdates:
		_, err = r.sendUpdates(requestCtx, body)
	default:
		return nil, ErrorMethodNotExists
	}
	if err != nil {
		if e, ok := status.FromError(err); ok {
			return NewRPCResponse(e.Code()), nil
		} else {
			return nil, err
		}
	}

	return NewRPCResponse(codes.OK), nil
}

// Close Закрытие подключения
func (r *RPCClient) Close() error {
	return r.conn.Close()
}

// createMeta создание метаинформации запроса
func (r RPCClient) createMeta(headers []Header) context.Context {
	md := make(map[string]string, len(headers))
	for _, h := range headers {
		md[h.Name] = h.Value
	}
	md["X-Real-IP"] = r.netAddr
	return metadata.NewOutgoingContext(r.ctx, metadata.New(md))
}

// sendUpdates отправка запроса на обновление метрик
func (r RPCClient) sendUpdates(ctx context.Context, body []byte) (*pb.MetricsResponse, error) {
	return r.service.HandleMetrics(ctx, &pb.MetricsRequest{
		Body: body,
	}, grpc.UseCompressor(gzip.Name))
}

// clearURL обработка урл сервера, отчистка от http
func clearURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = strings.Replace(url, "http://", "", 1)
	} else if strings.HasPrefix(url, "https://") {
		url = strings.Replace(url, "https://", "", 1)
	}
	return url
}

// EnableManualCompression Нужно ли сжимать тело запроса для этого клиента
func (r RPCClient) EnableManualCompression() bool {
	return false
}
