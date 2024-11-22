package sendpool

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRestClient(t *testing.T) {
	tests := []struct {
		desc    string
		baseURL string
		want    *resty.Client
	}{
		{
			desc:    "empty baseURL",
			baseURL: "",
			want:    &resty.Client{BaseURL: ""},
		},
		{
			desc:    "valid baseURL",
			baseURL: "https://example.com",
			want:    &resty.Client{BaseURL: "https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := NewRestClient(tt.baseURL)
			assert.NoError(t, err)
			if got.client.BaseURL != tt.want.BaseURL {
				t.Errorf("NewRestClient() baseURL = %v, want %v", got.client.BaseURL, tt.want.BaseURL)
			}
		})
	}
}

func TestRestClient_Post(t *testing.T) {
	tests := []struct {
		desc    string
		url     string
		body    []byte
		headers []Header
		err     error
	}{
		{
			desc:    "empty_body",
			url:     "",
			body:    []byte(""),
			headers: []Header{{Name: "Content-Type", Value: "application/json"}},
		},
		{
			desc:    "happy_path",
			url:     "",
			body:    []byte(`{"message":"Hello, World!"}`),
			headers: []Header{{Name: "Content-Type", Value: "application/json"}},
			err:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			router := chi.NewRouter()
			router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
				body, err := io.ReadAll(request.Body)
				if err != nil {
					t.Errorf("error reading body: %v", err)
					return
				}
				assert.Equal(t, tt.body, body)
				for _, header := range tt.headers {
					assert.Equal(t, header.Value, request.Header.Get(header.Name))
				}
				writer.WriteHeader(http.StatusOK)
				if _, wErr := writer.Write([]byte{}); wErr != nil {
					t.Errorf("error writing response: %v", wErr)
				}
			})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)

			client, cErr := NewRestClient(srv.URL)
			assert.NoError(t, cErr)
			_, err := client.Post(tt.url, tt.body, tt.headers...)
			assert.NoError(t, err)
		})
	}
}

// TestGetNetAddr tests the getNetAddr function.
func TestGetNetAddr(t *testing.T) {
	tests := []struct {
		desc    string
		wantErr bool
	}{
		{
			desc:    "valid IP address",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := getNetAddr()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}

func TestRESTClient_EnableManualCompression(t *testing.T) {
	tests := []struct {
		name      string
		client    RestClient
		expectVal bool
	}{
		{
			name:      "test_enable_manual_compression_false",
			client:    RestClient{},
			expectVal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.EnableManualCompression()
			assert.Equal(t, tt.expectVal, result, "EnableManualCompression() should return %v", tt.expectVal)
		})
	}
}
