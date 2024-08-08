package handlemetric

import (
	"encoding/json"
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
func JSONHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Log.Error(err)
		helpers.SetHTTPError(response, BadRequestError.HTTPStatus, helpers.GetErrorJSONBody(BadRequestError.Error()))
		return
	}
	// Парсим тело в структуру запроса
	var body payload.Metrics
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		logger.Log.Infow("Bad request for update metric", "error", err, "body", string(rawBody))
		helpers.SetHTTPError(response, BadRequestError.HTTPStatus, helpers.GetErrorJSONBody(BadRequestError.Error()))
		return
	}
	responseMessage, uError := updateMetricByRequestBody(body)
	if uError != nil {
		helpers.SetHTTPError(response, uError.HTTPStatus, helpers.GetErrorJSONBody(uError.Error()))
		return
	}
	rBody, rError := createResponse(body, responseMessage)
	if rError != nil {
		helpers.SetHTTPError(response, rError.HTTPStatus, helpers.GetErrorJSONBody(rError.Error()))
		return
	}
	response.WriteHeader(http.StatusOK)
	_, err = response.Write(rBody)
	if err != nil {
		logger.Log.Error(err)
	}
}

// createResponse создаём тело для ответа
func createResponse(body payload.Metrics, responseMessage string) ([]byte, *UpdateMetricError) {
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
