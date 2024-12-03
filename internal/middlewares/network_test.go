package middlewares

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilterNetwork(t *testing.T) {
	tt := []struct {
		testName    string
		subnet      string
		requestIP   string
		expectError error
	}{
		{
			testName:    "correct_ip",
			subnet:      "192.168.1.0/24",
			requestIP:   "192.168.1.2",
			expectError: nil,
		},
		{
			testName:    "ip_outside_subnet",
			subnet:      "192.168.1.0/24",
			requestIP:   "192.168.2.2",
			expectError: ErrorIPWrong,
		},
		{
			testName:    "ip_empty",
			subnet:      "192.168.1.0/24",
			requestIP:   "",
			expectError: ErrorIPEmpty,
		},
		{
			testName:    "nil_network",
			subnet:      "",
			requestIP:   "192.168.2.2",
			expectError: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			var network *net.IPNet
			var err error
			if tc.subnet != "" {
				_, network, err = net.ParseCIDR(tc.subnet)
				if err != nil {
					t.Fatal(err)
				}
			}
			middleware := NewNetworkMiddleware(network)

			router := chi.NewRouter()
			router.Use(middleware.FilterNetwork)
			router.Post("/", func(writer http.ResponseWriter, request *http.Request) {})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			request := resty.New().R()
			if tc.requestIP != "" {
				request.SetHeader("X-Real-IP", tc.requestIP)

			}
			request.Method = http.MethodPost
			request.URL = srv.URL
			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			if tc.expectError != nil {
				assert.Equal(t, http.StatusForbidden, res.StatusCode())
				assert.Contains(t, res.String(), tc.expectError.Error())
			} else {
				assert.Equal(t, http.StatusOK, res.StatusCode())
			}
		})
	}
}

func TestInterceptor(t *testing.T) {
	tt := []struct {
		testName    string
		subnet      string
		requestIP   string
		expectError error
		hasNoMD     bool
		ipNil       bool
	}{
		{
			testName:    "correct_ip",
			subnet:      "192.168.1.0/24",
			requestIP:   "192.168.1.2",
			expectError: nil,
		},
		{
			testName:    "ip_outside_subnet",
			subnet:      "192.168.1.0/24",
			requestIP:   "192.168.2.2",
			expectError: ErrorIPWrong,
		},
		{
			testName:    "ip_empty",
			subnet:      "192.168.1.0/24",
			requestIP:   "",
			expectError: ErrorIPEmpty,
		},
		{
			testName:    "ip_nil",
			subnet:      "192.168.1.0/24",
			requestIP:   "",
			expectError: ErrorIPEmpty,
			ipNil:       true,
		},
		{
			testName:    "nil_network",
			subnet:      "",
			requestIP:   "192.168.2.2",
			expectError: nil,
		},
		{
			testName:    "no_meta",
			subnet:      "192.168.1.0/24",
			requestIP:   "192.168.1.2",
			expectError: nil,
			hasNoMD:     true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			var network *net.IPNet
			var err error
			if tc.subnet != "" {
				_, network, err = net.ParseCIDR(tc.subnet)
				if err != nil {
					t.Fatal(err)
				}
			}
			middleware := NewNetworkMiddleware(network)

			var ctx context.Context
			var req = struct{}{} //fake request
			if tc.requestIP != "" {
				md := metadata.Pairs("X-Real-IP", tc.requestIP)
				ctx = metadata.NewIncomingContext(context.TODO(), md)
			} else {
				if tc.ipNil {
					md := metadata.Pairs("X-Real-IP1", "")
					ctx = metadata.NewIncomingContext(context.TODO(), md)
				} else {
					md := metadata.Pairs("X-Real-IP", "")
					ctx = metadata.NewIncomingContext(context.TODO(), md)
				}
			}
			if tc.hasNoMD {
				ctx = context.TODO()
			}
			resp, err := middleware.Interceptor(ctx, req, nil, func(ctx context.Context, req any) (any, error) { return nil, nil }) // passing nil for info and handler

			if tc.expectError != nil {
				assert.ErrorIs(t, err, tc.expectError)
			} else {
				assert.Equal(t, nil, resp)
				assert.NoError(t, err)
			}
		})
	}
}
