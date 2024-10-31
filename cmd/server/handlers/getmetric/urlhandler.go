package getmetric

import (
	"fmt"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// URLHandler Возвращает метрику
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP.
// - request: http.Request объект, содержащий информацию о запросе HTTP.
//
// @Summary Возвращает метрику
// @Description Данный хэндлер позволяет получить значение метрики по её типу и имени
// @Tags Метрики
// @Accept  json
// @Produce  json
// @Param type path string true "Тип метрики"
// @Param name path string true "Имя метрики"
// @Success 200 {string} string "значение метрики"
// @Failure 404 {string} string "метрика не найдена"
// @Router /value/{type}/{name} [get]
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
		if _, fErr := fmt.Fprint(response, value.ToString()); fErr != nil {
			logger.Log.Error(fErr)
		}
	case metrics.TypeCounter:
		value, ok := metrics.MeStore.GetCounter(metricName)
		if !ok {
			http.NotFound(response, request)
			return
		}
		if _, fErr := fmt.Fprint(response, value.ToString()); fErr != nil {
			logger.Log.Error(fErr)
		}
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
