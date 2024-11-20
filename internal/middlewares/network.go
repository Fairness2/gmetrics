package middlewares

import (
	"errors"
	"gmetrics/internal/helpers"
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

// FilterNetwork фильтрация запросов по подсети
func (nm *NetworkMiddleware) FilterNetwork(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if nm.network == nil {
			next.ServeHTTP(w, r)
			return
		}
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			helpers.SetHTTPResponse(w, http.StatusForbidden, helpers.GetErrorJSONBody(ErrorIPEmpty.Error()))
			return
		}
		if !nm.network.Contains(net.ParseIP(ip)) {
			helpers.SetHTTPResponse(w, http.StatusForbidden, helpers.GetErrorJSONBody(ErrorIPWrong.Error()))
			return
		}
		next.ServeHTTP(w, r)
	})
}
