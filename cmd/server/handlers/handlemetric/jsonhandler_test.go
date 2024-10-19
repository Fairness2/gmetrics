package handlemetric

import (
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
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
	}{
		{
			name:            "empty_type",
			body:            `{"id":"someName","type":"","value":123}`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "empty_name",
			body:            `{"id":"","type":"gauge","value":123}`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "empty_value",
			body:            `{"id":"someName","type":"gauge"}`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_type",
			body:            `{"id":"someName","type":"aboba","value":123}`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_value_gauge",
			body:            `{"id":"someName","type":"gauge","value":"some"}`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_value_count",
			body:            `{"id":"someName","type":"counter","delta":"some"}`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "right_value_gauge",
			body:            `{"id":"someName","type":"gauge","value":56.78}`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
		{
			name:            "right_value_count",
			body:            `{"id":"someName","type":"counter","delta":5}`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
	}
	router := chi.NewRouter()

	// Устанавливаем глобальное хранилище метрик
	storage := metrics.NewMemStorage()
	metrics.MeStore = storage

	router.Post("/update", JSONHandler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			request := resty.New().R()
			request.Method = http.MethodPost
			request.Body = test.body
			request.URL = srv.URL + "/update"

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.wantStatus, res.StatusCode(), "unexpected response status code")
		})
	}
}

func TestCreateResponse(t *testing.T) {
	tests := []struct {
		name      string
		body      func() payload.Metrics
		message   string
		want      []byte
		wantError bool
	}{
		{
			name: "gauge_type_with_existing_value",
			body: func() payload.Metrics {
				val := 12.34
				return payload.Metrics{ID: "someGauge", MType: metrics.TypeGauge, Value: &val}
			},
			message:   "gauge updated",
			want:      []byte(`{"status":"success","id":"someGauge","message":"gauge updated","value":12.34}`),
			wantError: false,
		},
		{
			name: "gauge_type_with_non_existing_value",
			body: func() payload.Metrics {
				return payload.Metrics{ID: "nonExistentGauge", MType: metrics.TypeGauge}
			},
			message:   "gauge updated",
			want:      []byte(`{"status":"success","id":"nonExistentGauge","message":"gauge updated"}`),
			wantError: false,
		},
		{
			name: "counter_type_with_existing_value",
			body: func() payload.Metrics {
				var val int64 = 123
				return payload.Metrics{ID: "someCounter", MType: metrics.TypeCounter, Delta: &val}
			},
			message:   "counter updated",
			want:      []byte(`{"status":"success","id":"someCounter","message":"counter updated","delta":123}`),
			wantError: false,
		},
		{
			name: "counter_type_with_non_existing_value",
			body: func() payload.Metrics {
				return payload.Metrics{ID: "nonExistentCounter", MType: metrics.TypeCounter}
			},
			message:   "counter updated",
			want:      []byte(`{"status":"success","id":"nonExistentCounter","message":"counter updated"}`),
			wantError: false,
		},
		{
			name: "unexpected_type",
			body: func() payload.Metrics {
				return payload.Metrics{ID: "someName", MType: "unexpectedType"}
			},
			message:   "unexpected type",
			want:      []byte(`{"status":"success","id":"someName","message":"unexpected type"}`),
			wantError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metrics.MeStore = metrics.NewMemStorage()
			body := test.body()
			switch body.MType {
			case metrics.TypeGauge:
				if body.Value != nil {
					if err := metrics.MeStore.SetGauge(body.ID, metrics.Gauge(*body.Value)); err != nil {
						t.Fatalf("error setting gauge: %v", err)
					}
				}
			case metrics.TypeCounter:
				if body.Delta != nil {
					if err := metrics.MeStore.AddCounter(body.ID, metrics.Counter(*body.Delta)); err != nil {
						t.Fatalf("error setting counter: %v", err)
					}
				}
			}
			got, err := createResponse(test.body(), test.message)
			assert.Equal(t, test.want, got, "unexpected response")
			if (err != nil) != test.wantError {
				t.Errorf("createResponse() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}
