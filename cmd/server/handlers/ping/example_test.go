package ping

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/http/httptest"
)

type dbExample struct{}

func (d dbExample) PingContext(ctx context.Context) error {
	return nil
}

// Example for Handler
func ExampleHandler() {
	// Set Server
	router := chi.NewRouter()
	router.Get("/", NewController(dbExample{}).Handler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// Set up an HTTP request.
	request := resty.New().R()
	request.Method = http.MethodGet
	request.URL = srv.URL + "/ping"

	_, _ = request.Send()
}
