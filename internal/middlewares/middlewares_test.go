package middlewares

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"gmetrics/cmd/server/config"
	pb "gmetrics/internal/payload/proto"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// hmacEncode создаём подпись запроса
func hmacEncode(key, content string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(content))

	return hex.EncodeToString(h.Sum(nil))
}

// TestCheckSign тест проверки подписи запроса
func TestCheckSignMiddleware(t *testing.T) {
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
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config.Params = &config.CliConfig{}
			if tc.hashKey != "" {
				config.Params.HashKey = tc.hashKey
			}

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

// TestJSONHeaders func tests the JSONHeaders function
func TestJSONHeaders(t *testing.T) {
	testCases := []struct {
		desc                string
		expectedContentType string
	}{
		{
			desc:                "json_content_type",
			expectedContentType: "application/json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := JSONHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

			handler.ServeHTTP(rr, req)

			result := rr.Header().Get("Content-Type")
			assert.Equal(t, tc.expectedContentType, result, "handler returned wrong Content-Type")
		})
	}
}

// TestGZIPDecompressRequest tests the GZIPDecompressRequest function
func TestGZIPDecompressRequest(t *testing.T) {
	testCases := []struct {
		desc         string
		body         string
		encoding     string
		expectedCode int
	}{
		{
			desc:         "gzip_encoded_body",
			body:         "gzip compressed body",
			encoding:     "gzip",
			expectedCode: http.StatusOK,
		},
		{
			desc:         "non_gzip_encoded_body",
			body:         "regular body",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			request := resty.New().R()

			var body []byte
			if tc.encoding != "" {
				var buf bytes.Buffer
				var err error
				zw := gzip.NewWriter(&buf)
				_, err = zw.Write([]byte(tc.body))
				if err != nil {
					t.Fatal(err)
				}
				if err := zw.Close(); err != nil {
					t.Fatal(err)
				}
				body = buf.Bytes()
				request.Header.Set("Content-Encoding", tc.encoding)
			} else {
				body = []byte(tc.body)
			}
			request.SetBody(body)

			router := chi.NewRouter()
			router.Use(GZIPDecompressRequest)
			router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
				rawBody, err := io.ReadAll(request.Body)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = writer.Write(rawBody); err != nil {
					t.Fatal(err)
				}
				writer.WriteHeader(http.StatusOK)
			})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			request.Method = http.MethodPost
			request.URL = srv.URL
			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")

			result := res.StatusCode()
			resultBody := res.Body()
			assert.Equal(t, tc.expectedCode, result, "handler returned wrong status code")
			assert.Equal(t, tc.body, string(resultBody), "handler returned wrong body")
		})
	}
}

// TestGZIPCompressResponse tests the GZIPCompressResponse function
func TestGZIPCompressResponse(t *testing.T) {
	testCases := []struct {
		desc         string
		body         string
		encoding     string
		expectedCode int
	}{
		{
			desc:         "gzip_encoded_body",
			body:         "gzip compressed body",
			encoding:     "gzip",
			expectedCode: http.StatusOK,
		},
		{
			desc:         "non_gzip_encoded_body",
			body:         "regular body",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			request := resty.New().R()
			request.Header.Set("Accept-Encoding", tc.encoding)

			router := chi.NewRouter()
			router.Use(GZIPCompressResponse)
			router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
				if _, wErr := writer.Write([]byte(tc.body)); wErr != nil {
					t.Fatal(wErr)
				}
				writer.WriteHeader(http.StatusOK)
			})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			request.Method = http.MethodPost
			request.URL = srv.URL
			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")

			result := res.StatusCode()
			resultBody := res.Body()

			h := res.Header().Get("Content-Encoding")
			assert.Equal(t, tc.encoding, h, "handler returned wrong Content-Encoding")

			/*var buf []byte
			if tc.encoding != "" {
				r, err := gzip.NewReader(bytes.NewReader(resultBody))
				if err != nil {
					t.Fatal(err)
				}
				if _, err := r.Read(buf); err != nil {
					t.Fatal(err)
				}
			} else {
				buf = res.Body()
			}*/

			assert.Equal(t, tc.expectedCode, result, "handler returned wrong status code")
			assert.Equal(t, tc.body, string(resultBody), "handler returned wrong body")

		})
	}
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
			desc:          "incorrect_body_hash",
			hashKey:       "key",
			hashHeader:    hmacEncode("key", "request body"),
			body:          "different body",
			expectedError: true,
		},
		{
			desc:          "incorrect_hash",
			hashKey:       "key",
			hashHeader:    "incorrect_hash",
			body:          "different body",
			expectedError: true,
		},
		{
			desc:          "missing_hash_key_in_config",
			hashHeader:    hmacEncode("key", "request body"),
			body:          "request body",
			expectedError: true,
		},
		{
			desc:          "missing_hash_in_header",
			hashKey:       "key",
			body:          "request body",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config.Params = &config.CliConfig{}
			if tc.hashKey != "" {
				config.Params.HashKey = tc.hashKey
			}

			err := checkSign(tc.hashHeader, []byte(tc.body))
			if !tc.expectedError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestCheckSign тест проверки подписи запроса
func TestCheckSignInterceptor(t *testing.T) {
	testCases := []struct {
		desc          string
		hashKey       string
		hashHeader    string
		body          string
		expectedError bool
		hashNil       bool
		hasNoMD       bool
	}{
		{
			desc:          "correct_hash",
			hashKey:       "key",
			hashHeader:    hmacEncode("key", "request body"),
			body:          "request body",
			expectedError: false,
		},
		{
			desc:          "incorrect_body_hash",
			hashKey:       "key",
			hashHeader:    hmacEncode("key", "request body"),
			body:          "different body",
			expectedError: true,
		},
		{
			desc:          "incorrect_hash",
			hashKey:       "key",
			hashHeader:    "incorrect_hash",
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
			expectedError: false,
		},
		{
			desc:          "missing_md",
			hashKey:       "key",
			body:          "request body",
			expectedError: false,
			hasNoMD:       true,
		},
		{
			desc:          "missing_md_hash",
			hashKey:       "key",
			body:          "request body",
			expectedError: false,
			hashNil:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			config.Params = &config.CliConfig{}
			if tc.hashKey != "" {
				config.Params.HashKey = tc.hashKey
			}

			var ctx context.Context
			var req = &pb.MetricsRequest{Body: []byte(tc.body)}
			if tc.hashHeader != "" {
				md := metadata.Pairs("HashSHA256", tc.hashHeader)
				ctx = metadata.NewIncomingContext(context.TODO(), md)
			} else {
				if tc.hashNil {
					md := metadata.Pairs("HashSHA2561", "")
					ctx = metadata.NewIncomingContext(context.TODO(), md)
				} else {
					md := metadata.Pairs("HashSHA256", "")
					ctx = metadata.NewIncomingContext(context.TODO(), md)
				}
			}
			if tc.hasNoMD {
				ctx = context.TODO()
			}
			_, err := CheckSignInterceptor(ctx, req, nil, func(ctx context.Context, req any) (any, error) { return nil, nil }) // passing nil for info and handler

			if !tc.expectedError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
