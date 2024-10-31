package getmetric

import (
	"gmetrics/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestJSONHandler(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		wantStatus      int
		wantContentType string
		wantValue       string
	}{
		{
			name:            "empty_type",
			body:            `{"id":"someName","type":""}`,
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantValue:       "",
		},
		{
			name:            "empty_name",
			body:            `{"id":"","type":"gauge"}`,
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantValue:       "",
		},
		{
			name:            "wrong_type",
			body:            `{"id":"someName","type":"aboba"}`,
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "not_empty_gauge",
			body:            `{"id":"someName","type":"gauge"}`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantValue:       `{"value":56.67,"id":"someName","type":"gauge"}`,
		},
		{
			name:            "not_empty_count",
			body:            `{"id":"someName","type":"counter"}`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantValue:       `{"delta":5,"id":"someName","type":"counter"}`,
		},
		{
			name:            "empty_gauge",
			body:            `{"id":"someName1","type":"gauge"}`,
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantValue:       "",
		},
		{
			name:            "empty_count",
			body:            `{"id":"someName1","type":"counter"}`,
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantValue:       "",
		},
	}
	router := chi.NewRouter()
	router.Post("/value", func(writer http.ResponseWriter, request *http.Request) {
		metrics.MeStore = metrics.NewMemStorage()
		_ = metrics.MeStore.SetGauge("someName", 56.67)
		_ = metrics.MeStore.AddCounter("someName", 5)
		JSONHandler(writer, request)
	})
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// останавливаем сервер после завершения теста

	defer srv.Close()
	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			request := resty.New().R()
			request.Method = http.MethodPost
			request.Body = test.body
			request.URL = srv.URL + "/value"

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.wantStatus, res.StatusCode(), "unexpected response status code")
			if test.wantStatus == http.StatusOK {
				assert.Equal(t, test.wantValue, string(res.Body()))
			}
		})
	}
}
