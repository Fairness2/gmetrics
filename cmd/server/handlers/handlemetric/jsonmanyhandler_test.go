package handlemetric

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONManyHandler(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		wantStatus      int
		wantContentType string
	}{
		{
			name:            "empty_type",
			body:            `[{"id":"someName","type":"","value":123}]`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "empty_name",
			body:            `[{"id":"","type":"gauge","value":123}]`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "empty_value",
			body:            `[{"id":"someName","type":"gauge"}]`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_type",
			body:            `[{"id":"someName","type":"aboba","value":123}]`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_value_gauge",
			body:            `[{"id":"someName","type":"gauge","value":"some"}]`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_value_count",
			body:            `[{"id":"someName","type":"counter","delta":"some"}]`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "right_value_gauge",
			body:            `[{"id":"someName","type":"gauge","value":56.78}]`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
		{
			name:            "right_value_count",
			body:            `[{"id":"someName","type":"counter","delta":5}]`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
	}
	router := chi.NewRouter()

	// Устанавливаем глобальное хранилище метрик
	storage := metrics.NewMemStorage()
	metrics.MeStore = storage

	router.Post("/updates", JSONManyHandler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			request := resty.New().R()
			request.Method = http.MethodPost
			request.Body = test.body
			request.URL = srv.URL + "/updates"

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.wantStatus, res.StatusCode(), "unexpected response status code")
		})
	}
}
