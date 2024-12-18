package getmetric

import (
	"encoding/json"
	"gmetrics/internal/helpers"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"io"
	"net/http"
)

// JSONHandler Возвращает метрику по запросу с JSON телом
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP
// - request: http.Request объект, содержащий информацию о запросе HTTP
//
// @Summary JSONHandler
// @Description Возвращает метрику по запросу с JSON телом
// @Tags Метрики
// @Accept json
// @Produce json
// @ID
// @Param request body payload.Metrics true "Metrics payload"
// @Success 200 {object} payload.Metrics
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /value [post]
func JSONHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Log.Error(err)
		helpers.SetHTTPResponse(response, http.StatusBadRequest, helpers.GetErrorJSONBody(err.Error()))
		return
	}
	// Парсим тело в структуру запроса
	var body payload.Metrics
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		logger.Log.Infow("Bad request for get metric", "error", err, "body", string(rawBody))
		helpers.SetHTTPResponse(response, http.StatusBadRequest, helpers.GetErrorJSONBody("Bad request for get metric"))
		return
	}
	switch body.MType {
	case metrics.TypeGauge:
		value, ok := metrics.MeStore.GetGauge(body.ID)
		if !ok {
			http.NotFound(response, request)
			return
		}
		rawValue := value.GetRaw()
		body.Value = &rawValue
	case metrics.TypeCounter:
		value, ok := metrics.MeStore.GetCounter(body.ID)
		if !ok {
			http.NotFound(response, request)
			return
		}
		rawValue := value.GetRaw()
		body.Delta = &rawValue
	default:
		http.NotFound(response, request)
		return
	}

	jsonResponse, err := json.Marshal(body)
	if err != nil {
		logger.Log.Infow("Bad request for get metric", "error", err, "body", string(rawBody))
		helpers.SetHTTPResponse(response, http.StatusInternalServerError, helpers.GetErrorJSONBody(err.Error()))
		return
	}
	response.WriteHeader(http.StatusOK)
	_, err = response.Write(jsonResponse)
	if err != nil {
		logger.Log.Error(err)
	}
}
