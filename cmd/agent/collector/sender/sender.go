package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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
	// В первый раз отправляем метрики через одну секунду, чтобы дать собрать метрики
	<-time.After(1 * time.Second)
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

	client := c.client.R().SetHeader("Content-Type", "application/json")

	// Пробуем сжать тело
	compressedBody, err := c.compressBody(body)
	if err != nil {
		// Если не получилось, то ставим обычное боди
		log.Println("Cant compress body", err)
		client.SetBody(body)
	} else {
		// Если получилось, то ставим заголовок о методе кодировки и ставим закодированное тело
		client.SetHeader("Content-Encoding", "gzip").
			SetBody(compressedBody)
	}

	// Отправляем запрос
	res, err := client.Post(sURL)
	log.Printf("Finish sending metric %s with value %d delta %d type %s\n", body.ID, body.Value, body.Delta, body.MType)
	if err != nil {
		return err
	}
	if statusCode := res.StatusCode(); statusCode != http.StatusOK {
		return fmt.Errorf("http status code %d", statusCode)
	}

	return nil
}

// compressBody сжимаем данные в формат gzip
func (c *Client) compressBody(body payload.Metrics) ([]byte, error) {
	// Преобразовываем тело в строку джейсон
	encodedBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Создаём буфер, в который запишем сжатое тело
	// TODO Тут мы создаём новый врайтер на каждый запрос, для выделения памяти это не особо хорошо,
	// TODO другой вариант создать его один раз в клиенте и использовать writer.Reset(&buf), но возникает проблема использования в нескольких потоках одновременно и нужно как то блокировать
	// TODO Какой вариант предпочтительнее или есть ещё варианты?
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	_, err = writer.Write(encodedBody)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
