package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/config"
	"gmetrics/internal/logger"
	"gmetrics/internal/metricerrors"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	client            *resty.Client // Клиент для подключения к серверам
	metricsCollection *collection.Type
	encodeWriterPool  sync.Pool
}

func New(mCollection *collection.Type) *Client {
	c := &Client{
		metricsCollection: mCollection,
		client:            resty.New(),
		encodeWriterPool: sync.Pool{
			New: func() interface{} {
				writer, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)
				if err != nil {
					logger.Log.Error(err)
					return nil
				}
				return writer
			},
		}}
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
	logger.Log.Info("Sending metrics")
	// urlTemplate Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update
	var urlUpdateTemplate = "%s/updates"
	sURL := fmt.Sprintf(urlUpdateTemplate, config.Params.ServerURL)

	client := c.client.R().SetHeader("Content-Type", "application/json")

	// Пробуем сжать тело
	compressedBody, err := c.compressBody(body)
	if err != nil {
		// Если не получилось, то ставим обычное боди
		logger.Log.Error("Cant compress body", err)
		client.SetBody(body)
	} else {
		// Если получилось, то ставим заголовок о методе кодировки и ставим закодированное тело
		client.SetHeader("Content-Encoding", "gzip").
			SetBody(compressedBody)
	}

	// Отправляем запрос
	res, err := client.Post(sURL)
	logger.Log.Info("Finish sending metrics")
	if err != nil {
		return metricerrors.NewRetriable(err)
	}
	if statusCode := res.StatusCode(); statusCode != http.StatusOK {
		return metricerrors.NewRetriable(fmt.Errorf("http status code %d", statusCode))
	}

	return nil
}

// compressBody сжимаем данные в формат gzip
func (c *Client) compressBody(body []payload.Metrics) ([]byte, error) {
	// Преобразовываем тело в строку джейсон
	encodedBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Создаём буфер, в который запишем сжатое тело
	var buf bytes.Buffer
	// Берём свободный врайтер для записи
	writer := c.encodeWriterPool.Get().(*gzip.Writer)
	// Если у нас он nil, то была ошибка
	if writer == nil {
		return nil, errors.New("writer is nil")
	}
	// Устанавливаем новый буфер во врайтер
	writer.Reset(&buf)
	// Кодируем данные
	_, err = writer.Write(encodedBody)
	if err != nil {
		return nil, err
	}
	// Делаем закрытие врайтера, чтобы он прописал все нужные байты в конце. После Reset он откроется снова (Проверял,ставил последовательно закрытие, ресет с новым буфером и снова запись)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
