package getmetrics

import (
	"gmetrics/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	type args struct {
		gauges   map[string]metrics.Gauge
		counters map[string]metrics.Counter
	}

	tests := []struct {
		name      string
		sendURL   string
		values    args
		wantValue map[string]string
	}{
		{
			name:      "no_metrics",
			values:    args{},
			wantValue: map[string]string{},
		},
		{
			name: "some_gauges",
			values: args{
				gauges: map[string]metrics.Gauge{
					"test_gauge1": 56.34,
					"test_gauge2": 34.66,
				},
			},
			wantValue: map[string]string{
				"test_gauge1": "56.34",
				"test_gauge2": "34.66",
			},
		},
		{
			name: "some_counters",
			values: args{
				counters: map[string]metrics.Counter{
					"test_counter1": 4,
					"test_counter2": 55,
				},
			},
			wantValue: map[string]string{
				"test_counter1": "4",
				"test_counter2": "55",
			},
		},
		{
			name: "gauges_and_counters",
			values: args{
				gauges: map[string]metrics.Gauge{
					"test_gauge1": 56.34,
					"test_gauge2": 34.66,
				},
				counters: map[string]metrics.Counter{
					"test_counter1": 4,
					"test_counter2": 55,
				},
			},
			wantValue: map[string]string{
				"test_gauge1":   "56.34",
				"test_gauge2":   "34.66",
				"test_counter1": "4",
				"test_counter2": "55",
			},
		},
	}

	stor := metrics.NewMemStorage()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
				stor.Gauge = tc.values.gauges
				stor.Counter = tc.values.counters
				metrics.MeStore = stor
				Handler(writer, request)
			})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			request := resty.New().R()
			request.Method = http.MethodGet
			request.URL = srv.URL
			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, http.StatusOK, res.StatusCode())
			for k, v := range tc.wantValue {
				assert.Contains(t, string(res.Body()), k, v)
			}

		})
	}
}
