package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"gmetrics/cmd/server/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

// hmacEncode создаём подпись запроса
func hmacEncode(key, content string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(content))

	return hex.EncodeToString(h.Sum(nil))
}

// TestCheckSign тест проверки подписи запроса
func TestCheckSign(t *testing.T) {
	testCases := []struct {
		desc          string
		hashKey       string
		hashHeader    string
		body          string
		expectedError bool
	}{
		{
			desc:          "correct_hash",
			hashKey:       "key",
			hashHeader:    hmacEncode("key", "request body"),
			body:          "request body",
			expectedError: false,
		},
		{
			desc:          "incorrect_hash",
			hashKey:       "key",
			hashHeader:    hmacEncode("key", "request body"),
			body:          "different body",
			expectedError: true,
		},
		{
			desc:          "missing_hash_key_in_config",
			hashHeader:    hmacEncode("key", "request body"),
			body:          "request body",
			expectedError: false,
		},
		{
			desc:          "missing_hash_in_header",
			hashKey:       "key",
			body:          "request body",
			expectedError: true,
		},
	}
	config.Params = &config.CliConfig{}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config.Params.HashKey = tc.hashKey

			router := chi.NewRouter()
			router.Use(CheckSign)
			router.Post("/", func(writer http.ResponseWriter, request *http.Request) {})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			request := resty.New().R()
			request.Header.Set("HashSHA256", tc.hashHeader)
			request.SetBody(tc.body)
			request.Method = http.MethodPost
			request.URL = srv.URL
			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			if !tc.expectedError {
				assert.Equal(t, http.StatusOK, res.StatusCode())
			} else {
				assert.Equal(t, http.StatusBadRequest, res.StatusCode())
			}
		})
	}
}
