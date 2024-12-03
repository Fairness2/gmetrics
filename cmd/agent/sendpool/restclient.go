package sendpool

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"net"
)

var ErrorNoIPAddres = errors.New("no ip address")

// Header структура для установки хэдера
type Header struct {
	Name  string
	Value string
}

// RestClient представляет собой клиент для отправки HTTP-запросов с использованием библиотеки Resty.
type RestClient struct {
	client  *resty.Client
	netAddr string // реальный адрес кликета, будет встроен в X-Real-IP
}

// Post отправляет HTTP POST запрос на заданный URL с указанным телом и опциональными заголовками.
func (r RestClient) Post(url string, body []byte, headers ...Header) (MetricResponse, error) {
	client := r.client.R()
	for _, header := range headers {
		client.SetHeader(header.Name, header.Value)
	}
	client.SetHeader("X-Real-IP", r.netAddr)
	client.SetBody(body)
	return client.Post(url)
}

// NewRestClient инициализирует новый RestClient с предоставленным базовым URL-адресом.
func NewRestClient(baseURL string) (*RestClient, error) {
	c := resty.New()
	c.BaseURL = baseURL
	addr, err := getNetAddr()
	if err != nil {
		return nil, err
	}
	return &RestClient{client: c, netAddr: addr}, nil
}

// getNetAddr находим среди адресов комьютера IPv4, который не является локальным
func getNetAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", ErrorNoIPAddres
}

// EnableManualCompression Необходимо ли вручную сжэимать тело запроса
func (r RestClient) EnableManualCompression() bool {
	return true
}
