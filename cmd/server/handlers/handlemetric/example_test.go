package handlemetric

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
	router.Post("/update/{type}/{name}/{value}", URLHandler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// Set up an HTTP request.
	request := resty.New().R()
	request.Method = http.MethodPost
	// Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	request.URL = srv.URL + "update/gauge/example/10.22"

	_, _ = request.Send()
}

// Example for JSONManyHandler
func ExampleJSONManyHandler() {
	// Set Server
	router := chi.NewRouter()
	router.Post("/updates", JSONManyHandler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// Set up an HTTP request.
	request := resty.New().R()
	request.Method = http.MethodPost
	request.SetBody(`[{"id":"some","type":"counter","delta":3},{"id":"some","type":"counter","delta":3},{"id":"some","type":"gauge","value":76.56}]`)
	request.URL = srv.URL + "/updates"

	_, _ = request.Send()
}

// Example for JSONHandler
func ExampleJSONHandler() {
	// Set Server
	router := chi.NewRouter()
	router.Post("/update", JSONHandler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(router)
	// Set up an HTTP request.
	request := resty.New().R()
	request.Method = http.MethodPost
	request.SetBody(`{"id":"some","type":"gauge","value":76.56}`)
	request.URL = srv.URL + "/updates"

	_, _ = request.Send()
}
