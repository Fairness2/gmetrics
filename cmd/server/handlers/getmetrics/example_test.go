package getmetrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/http/httptest"
)

// Example for Handler
func ExampleHandler() {
	// Set Server
	router := chi.NewRouter()
	router.Get("/", Handler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// Set up an HTTP request.
	request := resty.New().R()
	request.Method = http.MethodGet
	request.URL = srv.URL

	_, _ = request.Send()
}
