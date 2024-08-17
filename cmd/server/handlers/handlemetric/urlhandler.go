package handlemetric

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"gmetrics/internal/helpers"
	"net/http"
)

// URLHandler Обработка запроса установки метрики через параметры в строке урл
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP
// - request: http.Request объект, содержащий информацию о запросе HTTP
func URLHandler(response http.ResponseWriter, request *http.Request) {
	// Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	metricType, metricName, metricValue, err := parseURL(request)
	if err != nil {
		http.NotFound(response, request)
		return
	}

	updatedErr := updateMetricByStringValue(metricType, metricName, metricValue)
	if updatedErr != nil {
		var metricErr *UpdateMetricError
		if errors.As(updatedErr, &metricErr) {
			helpers.SetHTTPResponse(response, metricErr.HTTPStatus, helpers.GetErrorJSONBody(metricErr.Error()))
		} else {
			helpers.SetHTTPResponse(response, http.StatusInternalServerError, helpers.GetErrorJSONBody(updatedErr.Error()))
		}
		return
	}
	fmt.Fprintf(response, "metric %s successfully set", metricName)
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
