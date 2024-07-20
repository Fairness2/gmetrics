package sender

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/config"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"log"
	"net/http"
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
	// В первый раз отправляем метрики сразу же
	c.sendMetrics()
	for {
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-time.After(config.Params.ReportInterval):
			c.sendMetrics()
		case <-ctx.Done():
			log.Println("Periodic sender stopped")
			return
		}
	}
}

// sendMetrics Функция прохода по метрикам и запуск их отправки
func (c *Client) sendMetrics() {
	log.Println("Sending metrics")
	// Блокируем коллекцию на изменения
	c.metricsCollection.Lock()
	defer c.metricsCollection.Unlock()

	// Отправляем все собранные метрики
	for name, value := range c.metricsCollection.Values {
		switch value := value.(type) {
		case metrics.Gauge:
			metricValue := value.GetRaw()
			body := payload.Metrics{
				ID:    name,
				MType: metrics.TypeGauge,
				Value: &metricValue,
			}
			go func() {
				err := c.sendMetric(body)
				if err != nil {
					log.Println(err)
				}
			}()
		case metrics.Counter:
			metricValue := value.GetRaw()
			body := payload.Metrics{
				ID:    name,
				MType: metrics.TypeCounter,
				Delta: &metricValue,
			}
			go func() {
				err := c.sendMetric(body)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}
	// Отдельно отправляем каунт сбора метрик
	pCnt := c.metricsCollection.PollCount.GetRaw()
	body := payload.Metrics{
		ID:    "PollCount",
		MType: metrics.TypeCounter,
		Delta: &pCnt,
	}
	go func() {
		err := c.sendMetric(body)
		if err != nil {
			log.Println(err)
		}
	}()

	c.metricsCollection.ResetCounter()
}

// sendMetric Отправка метрики
func (c *Client) sendMetric(body payload.Metrics) error {
	log.Printf("Sending metric %s with value %d, delta %d type %s\n", body.ID, body.Value, body.Delta, body.MType)
	// urlTemplate Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update
	var urlUpdateTemplate = "%s/update"
	sURL := fmt.Sprintf(urlUpdateTemplate, config.Params.ServerURL)
	res, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(sURL)
	log.Printf("Finish sending metric %s with value %d delta %d type %s\n", body.ID, body.Value, body.Delta, body.MType)
	if err != nil {
		return err
	}
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
