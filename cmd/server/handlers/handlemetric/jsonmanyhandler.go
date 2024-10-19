package handlemetric

import (
	"encoding/json"
	"errors"
	"gmetrics/internal/helpers"
	"gmetrics/internal/logger"
	"gmetrics/internal/payload"
	"io"
	"net/http"
)

// JSONManyHandler Обработка запроса установки множества метрик через пост запрос с телом
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP
// - request: http.Request объект, содержащий информацию о запросе HTTP
//
// @Summary Обработка запроса установки множества метрик через пост запрос с телом
// @Description Обработка запроса установки множества метрик через пост запрос с телом
// @Tags		 Метрики
// @Accept json
// @Produce json
// @Param request body []payload.Metrics true "список метрик"
// @Success 200 {object} payload.ResponseBody "успешный ответ"
// @Failure 400 {object} payload.ResponseBody "ошибка запроса"
// @Failure 500 {object} payload.ResponseBody "внутренняя ошибка"
// @Router /updates [post]
func JSONManyHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Log.Error(err)
		helpers.SetHTTPResponse(response, BadRequestError.HTTPStatus, helpers.GetErrorJSONBody(BadRequestError.Error()))
		return
	}
	// Парсим тело в структуру запроса
	var body []payload.Metrics
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		logger.Log.Infow("Bad request for update metric", "error", err, "body", string(rawBody))
		helpers.SetHTTPResponse(response, BadRequestError.HTTPStatus, helpers.GetErrorJSONBody(BadRequestError.Error()))
		return
	}
	var metricErr *UpdateMetricError
	uError := updateMetricsByRequestBody(body)
	if uError != nil {
		if errors.As(uError, &metricErr) {
			helpers.SetHTTPResponse(response, metricErr.HTTPStatus, helpers.GetErrorJSONBody(metricErr.Error()))
		} else {
			helpers.SetHTTPResponse(response, http.StatusInternalServerError, helpers.GetErrorJSONBody(uError.Error()))
		}
		return
	}
	rBody, rError := createEmptyResponse("Metrics successfully updated.")
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

// createEmptyResponse создаём тело для ответа
func createEmptyResponse(responseMessage string) ([]byte, error) {
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
