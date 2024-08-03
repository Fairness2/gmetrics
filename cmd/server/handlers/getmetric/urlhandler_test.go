package getmetric

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
			input:     "/value/metrictype/metricname",
			wantType:  "metrictype",
			wantName:  "metricname",
			wantValue: "42",
			wantErr:   false,
		},
		{
			name:      "empty_section_URL",
			input:     "/update/metrictype//",
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
					gotType, gotName, err := parseURL(r)

					if tc.wantErr {
						assert.Error(t, err, "Not error parsing URL")
					} else {
						assert.NoError(t, err, "Unexpected error parsing URL")
						assert.Equal(t, tc.wantType, gotType)
						assert.Equal(t, tc.wantName, gotName)
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
			})
		})
	}
}

func TestURLHandler(t *testing.T) {
	// urlTemplate Шаблон урл: http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
	var urlUpdateTemplate = "/value/%s/%s"
	tests := []struct {
		name            string
		sendURL         string
		wantStatus      int
		wantContentType string
		wantValue       string
	}{
		{
			name:            "empty_type",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, "", "someName"),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantValue:       "",
		},
		{
			name:            "empty_name",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, ""),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantValue:       "",
		},
		{
			name:            "wrong_type",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, "aboba", "someName"),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "not_empty_gauge",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "someName"),
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantValue:       "56.67",
		},
		{
			name:            "not_empty_count",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeCounter, "someName"),
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantValue:       "5",
		},
		{
			name:            "empty_gauge",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "someName1"),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantValue:       "",
		},
		{
			name:            "empty_count",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeCounter, "someName1"),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantValue:       "",
		},
	}
	router := chi.NewRouter()
	router.Get("/value/{type}/{name}", func(writer http.ResponseWriter, request *http.Request) {
		metrics.MeStore = metrics.NewMemStorage()
		metrics.MeStore.SetGauge("someName", 56.67)
		metrics.MeStore.AddCounter("someName", 5)
		URLHandler(writer, request)
	})
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// останавливаем сервер после завершения теста

	defer srv.Close()
	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			request := resty.New().R()
			request.Method = http.MethodGet
			request.URL = srv.URL + test.sendURL

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.wantStatus, res.StatusCode())
			if test.wantStatus == http.StatusOK {
				assert.Equal(t, test.wantValue, string(res.Body()))
			}
		})
	}
}
