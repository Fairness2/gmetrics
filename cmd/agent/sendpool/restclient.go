package sendpool

import (
	"github.com/go-resty/resty/v2"
)

// Header структура для установки хэдера
type Header struct {
	Name  string
	Value string
}

// RestClient представляет собой клиент для отправки HTTP-запросов с использованием библиотеки Resty.
type RestClient struct {
	client *resty.Client
}

// Post отправляет HTTP POST запрос на заданный URL с указанным телом и опциональными заголовками.
func (r RestClient) Post(url string, body []byte, headers ...Header) (*resty.Response, error) {
	client := r.client.R()
	for _, header := range headers {
		client.SetHeader(header.Name, header.Value)
	}
	client.SetBody(body)
	return client.Post(url)
}

// NewRestClient инициализирует новый RestClient с предоставленным базовым URL-адресом.
func NewRestClient(baseURL string) *RestClient {
	c := resty.New()
	c.BaseURL = baseURL
	return &RestClient{client: c}
}
