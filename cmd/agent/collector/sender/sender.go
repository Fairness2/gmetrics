package sender

import (
	"context"
	"errors"
	"fmt"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/config"
	"gmetrics/cmd/agent/sendpool"
	"gmetrics/internal/logger"
	"gmetrics/internal/metricerrors"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client представляет собой клиент для подключения к серверам и отправки метрик.
type Client struct {
	client            *resty.Client // Клиент для подключения к серверам
	metricsCollection *collection.Type
	sendPool          Sender
}

// Sender интерфейс для пула конектов к серверу
type Sender interface {
	Send(body []payload.Metrics) (sendpool.MetricResponse, error)
}

// New инициализирует и возвращает новый экземпляр клиента с заданным набором метрик и пулом отправки.
func New(mCollection *collection.Type, sendPool Sender) *Client {
	c := &Client{
		metricsCollection: mCollection,
		client:            resty.New(),
		sendPool:          sendPool,
	}
	return c
}

// PeriodicSender Циклическая отправка данных
func (c *Client) PeriodicSender(ctx context.Context) {
	logger.Log.Info("Starting periodic sender")
	ticker := time.NewTicker(time.Duration(config.Params.ReportInterval) * time.Second)
	c.retrySend()
	for {
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-ticker.C:
			c.retrySend()
		case <-ctx.Done():
			logger.Log.Info("Periodic sender stopped")
			ticker.Stop()
			return
		}
	}
}

// retrySend отправка метрик с повторами
func (c *Client) retrySend() {
	pause := time.Second
	var rErr *metricerrors.Retriable
	for i := 0; i < 3; i++ {
		err := c.sendMetrics()
		if err == nil {
			break
		}
		logger.Log.Error(err)
		if !errors.As(err, &rErr) {
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
}

// sendMetrics Функция прохода по метрикам и запуск их отправки
func (c *Client) sendMetrics() error {
	logger.Log.Info("Sending metrics")
	// Блокируем коллекцию на изменения
	c.metricsCollection.Lock()
	defer c.metricsCollection.Unlock()
	body := make([]payload.Metrics, 0, len(c.metricsCollection.Values))

	// Отправляем все собранные метрики
	for name, value := range c.metricsCollection.Values {
		switch value := value.(type) {
		case metrics.Gauge:
			metricValue := value.GetRaw()
			body = append(body, payload.Metrics{
				ID:    name,
				MType: metrics.TypeGauge,
				Value: &metricValue,
			})
		case metrics.Counter:
			metricValue := value.GetRaw()
			body = append(body, payload.Metrics{
				ID:    name,
				MType: metrics.TypeCounter,
				Delta: &metricValue,
			})
		}
	}
	// Отдельно отправляем каунт сбора метрик
	pCnt := c.metricsCollection.PollCount.GetRaw()
	body = append(body, payload.Metrics{
		ID:    "PollCount",
		MType: metrics.TypeCounter,
		Delta: &pCnt,
	})
	// Отправляем метрики, но сбрасываем каунтер только при успешной отправке
	if err := c.sendToServer(body); err != nil {
		return err
	}
	c.metricsCollection.ResetCounter()

	return nil
}

// sendToServer Отправка метрики
func (c *Client) sendToServer(body []payload.Metrics) error {
	// Отправляем запрос
	res, err := c.sendPool.Send(body)
	logger.Log.Info("Finish sending metrics")
	if err != nil {
		//return metricerrors.NewRetriable(err)
		return err
	}
	if statusCode := res.StatusCode(); statusCode != http.StatusOK {
		return metricerrors.NewRetriable(fmt.Errorf("http status code %d", statusCode))
	}

	return nil
}
