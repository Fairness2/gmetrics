package handlers

import (
	"fmt"
	"gmetrics/internal/metrics"
	"net/http"
)

// GetMetrics Возвращает хранилище метрик. Метод для отладки
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP.
// - request: http.Request объект, содержащий информацию о запросе HTTP.
func GetMetrics(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(response, "%v", metrics.MeStore)
}
