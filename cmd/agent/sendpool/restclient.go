package sendpool

import (
	"github.com/go-resty/resty/v2"
)

type Header struct {
	Name  string
	Value string
}

type RestClient struct {
	client *resty.Client
}

func (r RestClient) Post(url string, body []byte, headers ...Header) (*resty.Response, error) {
	client := r.client.R()
	for _, header := range headers {
		client.SetHeader(header.Name, header.Value)
	}
	client.SetBody(body)
	return client.Post(url)
}

func NewRestClient(baseURL string) *RestClient {
	c := resty.New()
	c.BaseURL = baseURL
	return &RestClient{client: c}
}
