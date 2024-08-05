package getmetric

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"gmetrics/internal/metrics"
	"net/http"
)

// URLHandler Возвращает метрику
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP.
// - request: http.Request объект, содержащий информацию о запросе HTTP.
func URLHandler(response http.ResponseWriter, request *http.Request) {
	metricType, metricName, err := parseURL(request)
	if err != nil {
		http.NotFound(response, request)
		return
	}

	switch metricType {
	case metrics.TypeGauge:
		value, ok := metrics.MeStore.GetGauge(metricName)
		if !ok {
			http.NotFound(response, request)
			return
		}
		fmt.Fprint(response, value.ToString())
	case metrics.TypeCounter:
		value, ok := metrics.MeStore.GetCounter(metricName)
		if !ok {
			http.NotFound(response, request)
			return
		}
		fmt.Fprint(response, value.ToString())
	default:
		http.NotFound(response, request)
		return
	}
}

// parseURL Разбор URL на тип метрики, имя метрики
// Parameters:
// - request
// Returns:
// - metricType: тип метрики
// - metricName: имя метрики
// - error: ошибка при неверном URL
func parseURL(request *http.Request) (string, string, error) {
	metricType := chi.URLParam(request, "type")
	metricName := chi.URLParam(request, "name")

	if metricType == "" || metricName == "" {
		return "", "", fmt.Errorf("invalid URL: %s", request.URL.Path)
	}

	return metricType, metricName, nil
}
