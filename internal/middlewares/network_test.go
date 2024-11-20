package middlewares

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
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
