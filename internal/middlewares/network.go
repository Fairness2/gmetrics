package middlewares

import (
	"context"
	"errors"
	"gmetrics/internal/helpers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
)

var (
	ErrorIPEmpty = errors.New("ip is empty")
	ErrorIPWrong = errors.New("ip is wrong")
)

type NetworkMiddleware struct {
	network *net.IPNet
}

func NewNetworkMiddleware(subnet *net.IPNet) *NetworkMiddleware {
	return &NetworkMiddleware{network: subnet}
}

// FilterNetwork фильтрация запросов по подсети. Мидлваре
func (nm *NetworkMiddleware) FilterNetwork(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if nm.network == nil {
			next.ServeHTTP(w, r)
			return
		}
		ip := r.Header.Get("X-Real-IP")
		err := nm.checkIP(ip)
		if err != nil {
			helpers.SetHTTPResponse(w, http.StatusForbidden, helpers.GetErrorJSONBody(err.Error()))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// checkIP фильтрация запросов по подсети
func (nm *NetworkMiddleware) checkIP(ip string) error {
	if ip == "" {
		return ErrorIPEmpty
	}
	if !nm.network.Contains(net.ParseIP(ip)) {
		return ErrorIPWrong
	}
	return nil
}

// Interceptor фильтрация запросов по подсети для rpc
func (nm *NetworkMiddleware) Interceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if nm.network == nil {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}
	var ip string
	h := md.Get("X-Real-IP")
	if len(h) > 0 {
		ip = h[0]
	}
	err := nm.checkIP(ip)
	if err != nil {
		return nil, errors.Join(status.Error(codes.PermissionDenied, "ip is not prohibited"), err)
	}
	return handler(ctx, req)
}
