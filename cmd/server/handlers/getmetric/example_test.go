package getmetric

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/http/httptest"
)

// Example for URLHandler
func ExampleURLHandler() {
	// Set Server
	router := chi.NewRouter()
	router.Get("/value/{type}/{name}", URLHandler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// Set up an HTTP request.
	request := resty.New().R()
	request.Method = http.MethodGet
	// Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
	request.URL = srv.URL + "/value/gauge/example_metric"

	_, _ = request.Send()
}

// Example for JSONHandler
func ExampleJSONHandler() {
	// Set Server
	router := chi.NewRouter()
	router.Post("/value", URLHandler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// Set up an HTTP request.
	request := resty.New().R()
	request.Method = http.MethodPost
	request.SetBody(`{"id":"some","type":"counter"}`)
	request.URL = srv.URL + "/value"

	_, _ = request.Send()
}
