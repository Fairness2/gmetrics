package sendpool

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
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
			got := NewRestClient(tt.baseURL)
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

			client := NewRestClient(srv.URL)
			_, err := client.Post(tt.url, tt.body, tt.headers...)
			assert.NoError(t, err)
		})
	}
}
