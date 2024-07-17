package sender

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/config"
	"gmetrics/internal/metrics"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client            *resty.Client // Клиент для подключения к серверам
	metricsCollection *collection.Type
}

func New(mCollection *collection.Type) *Client {
	c := &Client{
		metricsCollection: mCollection,
		client:            resty.New(),
	}
	return c
}

// PeriodicSender Циклическая отправка данных
func (c *Client) PeriodicSender(ctx context.Context) {
	log.Println("Starting periodic sender")
	for {
		c.sendMetrics()
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-ctx.Done():
			log.Println("Periodic sender stopped")
			return
		default:
			//continue
		}
		time.Sleep(config.Params.ReportInterval)
	}
}

// sendMetrics Функция прохода по метрикам и запуск их отправки
func (c *Client) sendMetrics() {
	log.Println("Sending metrics")
	c.metricsCollection.Lock()
	defer c.metricsCollection.Unlock()
	// Отправляем все собранные метрики
	for name, value := range c.metricsCollection.Values {
		switch value := value.(type) {
		case metrics.Gauge:
			metricValue := value.ToString()
			go func() {
				err := c.sendMetric(metrics.TypeGauge, name, metricValue)
				if err != nil {
					log.Println(err)
				}
			}()
		case metrics.Counter:
			metricValue := value.ToString()
			go func() {
				err := c.sendMetric(metrics.TypeCounter, name, metricValue)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}
	// Отдельно отправляем каунт сбора метрик
	pCnt := c.metricsCollection.PollCount.ToString()
	go func() {
		err := c.sendMetric(metrics.TypeCounter, "PollCount", pCnt)
		if err != nil {
			log.Println(err)
		}
	}()

	c.metricsCollection.ResetCounter()
}

// sendMetric Отправка метрики
func (c *Client) sendMetric(mType string, name string, value string) error {
	log.Printf("Sending metric %s with value %s type %s\n", name, value, mType)
	// urlTemplate Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var urlUpdateTemplate = "%s/update/%s/%s/%s"
	sType := url.QueryEscape(mType)
	sName := url.QueryEscape(name)
	sValue := url.QueryEscape(value)
	sURL := fmt.Sprintf(urlUpdateTemplate, config.Params.ServerURL, sType, sName, sValue)
	res, err := c.client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(nil).
		Post(sURL)
	// res, err := c.client.Post(sURL, "text/plain", nil)
	log.Printf("Finish sending metric %s with value %s type %s\n", name, value, mType)
	if err != nil {
		return err
	}
	// defer res.Body.Close()
	if statusCode := res.StatusCode(); statusCode != http.StatusOK {
		return fmt.Errorf("http status code %d", statusCode)
	}

	return nil
}

// NewSender starts the process of sending metrics by creating a new sender client and
// launching a goroutine to execute the periodicSender function.
func NewSender(coll *collection.Type) *Client {
	client := New(coll)
	return client
}
