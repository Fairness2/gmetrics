package sender

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/config"
	"gmetrics/internal/metrics"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

type Client struct {
	client            *resty.Client // Клиент для подключения к серверам
	metricsCollection *collection.CollectionType
}

func New(mCollection *collection.CollectionType) *Client {
	c := &Client{
		metricsCollection: mCollection,
		client:            resty.New(),
	}
	return c
}

// StartSender Старт начала прослушивания и создание клиента
func (c *Client) StartSender() {
	go c.periodicSender()
}

// periodicSender Циклическая отправка данных
func (c *Client) periodicSender() {
	fmt.Println("Starting periodic sender")
	for {
		c.sendMetrics()
		time.Sleep(config.ReportInterval)
	}
}

// sendMetrics Функция прохода по метрикам и запуск их отправки
func (c *Client) sendMetrics() {
	fmt.Println("Sending metrics")
	c.metricsCollection.LockRead()
	defer c.metricsCollection.UnlockRead()
	v := reflect.Indirect(reflect.ValueOf(c.metricsCollection))
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)
		switch field.Interface().(type) {
		case metrics.Gauge:
			metricName := fieldType.Name
			metricValue := field.Interface().(metrics.Gauge).ToString()
			go func() {
				err := c.sendMetric(metrics.TypeGauge, metricName, metricValue)
				if err != nil {
					fmt.Println(err)
				}
			}()
		case metrics.Counter:
			metricName := fieldType.Name
			metricValue := field.Interface().(metrics.Counter).ToString()
			go func() {
				err := c.sendMetric(metrics.TypeCounter, metricName, metricValue)
				if err != nil {
					fmt.Println(err)
				}
			}()
		}
	}
	c.metricsCollection.ResetCounter()
}

// sendMetric Отправка метрики
func (c *Client) sendMetric(mType string, name string, value string) error {
	fmt.Printf("Sending metric %s with value %s type %s\n", name, value, mType)
	// urlTemplate Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var urlUpdateTemplate = "%s/update/%s/%s/%s"
	sType := url.QueryEscape(mType)
	sName := url.QueryEscape(name)
	sValue := url.QueryEscape(value)
	sURL := fmt.Sprintf(urlUpdateTemplate, config.ServerURL, sType, sName, sValue)
	res, err := c.client.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(nil).
		Post(sURL)
	// res, err := c.client.Post(sURL, "text/plain", nil)
	fmt.Printf("Finish sending metric %s with value %s type %s\n", name, value, mType)
	if err != nil {
		return err
	}
	// defer res.Body.Close()
	if statusCode := res.StatusCode(); statusCode != http.StatusOK {
		return fmt.Errorf("http status code %d", statusCode)
	}

	return nil
}
