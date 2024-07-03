package handleMetric

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"gmetrics/internal/helpers"
	"gmetrics/internal/metrics"
	"net/http"
	"strconv"
)

// Handler Обработка запроса установки метрики
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP
// - request: http.Request объект, содержащий информацию о запросе HTTP
func Handler(response http.ResponseWriter, request *http.Request) {
	// Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	metricType, metricName, metricValue, err := parseURL(request)
	if err != nil {
		http.NotFound(response, request)
		return
	}

	switch metricType {
	case metrics.TypeGauge:
		convertedValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			//log.Println(err)
			helpers.SetHTTPError(response, http.StatusBadRequest, "metric value is not a valid float")
			return
		}
		metrics.MeStore.SetGauge(metricName, metrics.Gauge(convertedValue))
		fmt.Fprintf(response, "metric %s successfully set", metricName)
	case metrics.TypeCounter:
		convertedValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			//log.Println(err)
			helpers.SetHTTPError(response, http.StatusBadRequest, "metric value is not a valid int")
			return
		}
		metrics.MeStore.AddCounter(metricName, metrics.Counter(convertedValue))
		fmt.Fprintf(response, "metric %s successfully add", metricName)
	default:
		helpers.SetHTTPError(response, http.StatusBadRequest, "invalid metric type")
		return
	}
}

// parseURL Разбор URL на тип метрики, имя метрики и значение метрики
// Parameters:
// - request
// Returns:
// - metricType: тип метрики
// - metricName: имя метрики
// - metricValue: значение метрики
// - error: ошибка при неверном URL
func parseURL(request *http.Request) (string, string, string, error) {
	metricType := chi.URLParam(request, "type")
	metricName := chi.URLParam(request, "name")
	metricValue := chi.URLParam(request, "value")

	if metricType == "" || metricName == "" || metricValue == "" {
		return "", "", "", fmt.Errorf("invalid URL: %s", request.URL.Path)
	}

	return metricType, metricName, metricValue, nil
}
