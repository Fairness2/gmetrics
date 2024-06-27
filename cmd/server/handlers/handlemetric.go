package handlers

import (
	"fmt"
	"gmetrics/internal/metrics"
	"net/http"
	"strconv"
	"strings"
)

// HandleMetric Обработка запроса установки метрики
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP
// - request: http.Request объект, содержащий информацию о запросе HTTP
func HandleMetric(response http.ResponseWriter, request *http.Request) {
	// Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	metricType, metricName, metricValue, err := parseURL(request.URL.Path)
	if err != nil {
		http.NotFound(response, request)
		return
	}

	switch metricType {
	case "gauge":
		convertedValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			//log.Println(err)
			setHTTPError(response, http.StatusBadRequest, "metric value is not a valid float")
			return
		}
		metrics.MeStore.SetGauge(metricName, metrics.Gauge(convertedValue))
		fmt.Fprintf(response, "metric %s successfully set", metricName)
	case "counter":
		convertedValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			//log.Println(err)
			setHTTPError(response, http.StatusBadRequest, "metric value is not a valid int")
			return
		}
		metrics.MeStore.AddCounter(metricName, metrics.Counter(convertedValue))
		fmt.Fprintf(response, "metric %s successfully add", metricName)
	default:
		setHTTPError(response, http.StatusBadRequest, "invalid metric type")
		return
	}
}

// setHTTPError Отправка ошибки и сообщения ошибки.
// Parameters:
// - response: http.ResponseWriter object containing information about the HTTP response
// - status: the HTTP status code to set in the response
// - message: the message to write to the response
func setHTTPError(response http.ResponseWriter, status int, message string) {
	response.WriteHeader(status)
	fmt.Fprint(response, message)
}

// parseURL Разбор URL на тип метрики, имя метрики и значение метрики
// Parameters:
// - url: URL, который нужно разобрать на части
// Returns:
// - metricType: тип метрики
// - metricName: имя метрики
// - metricValue: значение метрики
// - error: ошибка при неверном URL
func parseURL(url string) (string, string, string, error) {
	parts := strings.Split(url, "/")
	if len(parts) != 5 {
		return "", "", "", fmt.Errorf("invalid URL: %s", url)
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	if metricType == "" || metricName == "" || metricValue == "" {
		return "", "", "", fmt.Errorf("invalid URL: %s", url)
	}

	return metricType, metricName, metricValue, nil
}
