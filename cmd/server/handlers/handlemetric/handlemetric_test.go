package handlemetric

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseURL(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wantType  string
		wantName  string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid_URL",
			input:     "/update/metrictype/metricname/metricvalue",
			wantType:  "metrictype",
			wantName:  "metricname",
			wantValue: "metricvalue",
			wantErr:   false,
		},
		{
			name:      "incomplete_URL",
			input:     "/update/metrictype/metricname/",
			wantType:  "",
			wantName:  "",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "empty_section_URL",
			input:     "/update/metrictype//metricvalue",
			wantType:  "",
			wantName:  "",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "all_empty_URL",
			input:     "/update///",
			wantType:  "",
			wantName:  "",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "more_sections_URL",
			input:     "/update/metrictype/metricname/metricvalue/extra",
			wantType:  "",
			wantName:  "",
			wantValue: "",
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				router := chi.NewRouter()
				router.Get("/update/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
					gotType, gotName, gotValue, err := parseURL(r)

					if tc.wantErr {
						assert.Error(t, err, "Not error parsing URL")
					} else {
						assert.NoError(t, err, "Unexpected error parsing URL")
						assert.Equal(t, tc.wantType, gotType)
						assert.Equal(t, tc.wantName, gotName)
						assert.Equal(t, tc.wantValue, gotValue)
					}
				})
				// запускаем тестовый сервер, будет выбран первый свободный порт
				srv := httptest.NewServer(router)
				// останавливаем сервер после завершения теста
				defer srv.Close()
				// делаем запрос с помощью библиотеки resty к адресу запущенного сервера,
				// который хранится в поле URL соответствующей структуры
				req := resty.New().R()
				req.Method = http.MethodGet
				req.URL = srv.URL + tc.input
				_, err := req.Send()
				assert.NoError(t, err, "Unexpected error while sending request")
				/*if tc.wantErr {
					assert.NotEqual(t, res.StatusCode(), http.StatusOK)
				} else {
					assert.Equal(t, res.StatusCode(), http.StatusOK)
				}*/
			})
		})
	}
}

func TestHandleMetric(t *testing.T) {
	// urlTemplate Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var urlUpdateTemplate = "/update/%s/%s/%s"
	tests := []struct {
		name            string
		sendURL         string
		wantStatus      int
		wantContentType string
	}{
		{
			name:            "wrong_URL",
			sendURL:         "/update",
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "empty_type",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, "", "someName", "123"),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "empty_name",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "", "123"),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "empty_value",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "someName", ""),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_type",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, "aboba", "someName", "123"),
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_value_gauge",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "someName", "some"),
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "wrong_value_count",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeCounter, "someName", "some"),
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "right_value_gauge",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "someName", "56.78"),
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
		{
			name:            "right_value_count",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeCounter, "someName", "5"),
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
	}
	router := chi.NewRouter()

	// Устанавливаем глобальное хранилище метрик
	storage := metrics.NewMemStorage()
	metrics.MeStore = storage

	router.Post("/update/{type}/{name}/{value}", Handler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			request := resty.New().R()
			request.Method = http.MethodPost
			request.URL = srv.URL + test.sendURL

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.wantStatus, res.StatusCode())
		})
	}
}
