package handlemetric

import (
	"encoding/json"
	"errors"
	"fmt"
	"gmetrics/internal/helpers"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"io"
	"net/http"
)

// JSONHandler Обработка запроса установки метрики через пост запрос с телом
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP
// - request: http.Request объект, содержащий информацию о запросе HTTP
// @Summary Обработка запроса установки метрики через пост запрос с телом
// @Description Обработка запроса установки метрики через пост запрос с телом
// @Accept json
// @Produce json
// @Tags		 Метрики
// @Param metric body payload.Metrics true "Метрика"
// @Success 200 {object} payload.ResponseBody "Метрика успешно добавлена"
// @Failure 400 {object} payload.ErrorResponse "Неправильный запрос"
// @Failure 500 {object} payload.ErrorResponse "Внутренняя ошибка сервера"
// @Router /update [post]
func JSONHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Log.Error(err)
		helpers.SetHTTPResponse(response, BadRequestError.HTTPStatus, helpers.GetErrorJSONBody(BadRequestError.Error()))
		return
	}
	// Парсим тело в структуру запроса
	var body payload.Metrics
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		logger.Log.Infow("Bad request for update metric", "error", err, "body", string(rawBody))
		helpers.SetHTTPResponse(response, BadRequestError.HTTPStatus, helpers.GetErrorJSONBody(BadRequestError.Error()))
		return
	}
	var metricErr *UpdateMetricError
	uError := updateMetricByRequestBody(body)
	if uError != nil {
		if errors.As(uError, &metricErr) {
			helpers.SetHTTPResponse(response, metricErr.HTTPStatus, helpers.GetErrorJSONBody(metricErr.Error()))
		} else {
			helpers.SetHTTPResponse(response, http.StatusInternalServerError, helpers.GetErrorJSONBody(uError.Error()))
		}
		return
	}
	rBody, rError := createResponse(body, fmt.Sprintf("metric %s successfully add", body.ID))
	if rError != nil {
		if errors.As(rError, &metricErr) {
			helpers.SetHTTPResponse(response, metricErr.HTTPStatus, helpers.GetErrorJSONBody(metricErr.Error()))
		} else {
			helpers.SetHTTPResponse(response, http.StatusInternalServerError, helpers.GetErrorJSONBody(rError.Error()))
		}
		return
	}
	response.WriteHeader(http.StatusOK)
	_, err = response.Write(rBody)
	if err != nil {
		logger.Log.Error(err)
	}
}

// createResponse создаём тело для ответа
func createResponse(body payload.Metrics, responseMessage string) ([]byte, error) {
	rBody := payload.ResponseBody{
		Status:  payload.ResponseSuccessStatus,
		ID:      body.ID,
		Message: responseMessage,
	}
	switch body.MType {
	case metrics.TypeGauge:
		val, ok := metrics.MeStore.GetGauge(body.ID)
		if ok {
			rBody.Value = val.GetRaw()
		}
	case metrics.TypeCounter:
		val, ok := metrics.MeStore.GetCounter(body.ID)
		if ok {
			rBody.Delta = val.GetRaw()
		}
	}

	jsonResponse, err := json.Marshal(rBody)
	if err != nil {
		return []byte{}, &UpdateMetricError{
			error:      err,
			HTTPStatus: http.StatusInternalServerError,
		}
	}
	return jsonResponse, nil
}
