package getmetrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockMeStore struct {
	gauges   map[string]metrics.Gauge
	counters map[string]metrics.Counter
}

type MockMetrics struct{}

func (ms *MockMeStore) GetGauges() map[string]metrics.Gauge {
	return ms.gauges
}

func (ms *MockMeStore) GetCounters() map[string]metrics.Counter {
	return ms.counters
}

func (ms *MockMeStore) GetGauge(name string) (metrics.Gauge, bool) {
	return 5, true
}

func (ms *MockMeStore) GetCounter(name string) (metrics.Counter, bool) {
	return 5, true
}

func (ms *MockMeStore) SetGauge(name string, value metrics.Gauge) {

}

func (ms *MockMeStore) AddCounter(name string, value metrics.Counter) {

}

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
			name:      "no metrics",
			values:    args{},
			wantValue: map[string]string{},
		},
		{
			name: "some gauges",
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
			name: "some counters",
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
			name: "gauges and counters",
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
				metrics.MeStore = &MockMeStore{
					gauges:   tc.values.gauges,
					counters: tc.values.counters,
				}
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
