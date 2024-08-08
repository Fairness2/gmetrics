package handlemetric

import (
	"encoding/json"
	"gmetrics/internal/helpers"
	"gmetrics/internal/logger"
	"gmetrics/internal/payload"
	"io"
	"net/http"
)

// JSONHandler Обработка запроса установки множества метрик через пост запрос с телом
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP
// - request: http.Request объект, содержащий информацию о запросе HTTP
func JSONManyHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Log.Error(err)
		helpers.SetHTTPError(response, BadRequestError.HTTPStatus, helpers.GetErrorJSONBody(BadRequestError.Error()))
		return
	}
	// Парсим тело в структуру запроса
	var body []payload.Metrics
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		logger.Log.Infow("Bad request for update metric", "error", err, "body", string(rawBody))
		helpers.SetHTTPError(response, BadRequestError.HTTPStatus, helpers.GetErrorJSONBody(BadRequestError.Error()))
		return
	}
	responseMessage, uError := updateMetricsByRequestBody(body)
	if uError != nil {
		helpers.SetHTTPError(response, uError.HTTPStatus, helpers.GetErrorJSONBody(uError.Error()))
		return
	}
	rBody, rError := createEmptyResponse(responseMessage)
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

// createEmptyResponse создаём тело для ответа
func createEmptyResponse(responseMessage string) ([]byte, *UpdateMetricError) {
	rBody := payload.ResponseBody{
		Status:  payload.ResponseSuccessStatus,
		Message: responseMessage,
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
