package helpers

import (
	"encoding/json"
	"gmetrics/internal/logger"
	"gmetrics/internal/payload"
	"net/http"
)

// SetHTTPError Отправка ошибки и сообщения ошибки.
// Parameters:
// - response: http.ResponseWriter object containing information about the HTTP response
// - status: the HTTP status code to set in the response
// - message: the message to write to the response
func SetHTTPError(response http.ResponseWriter, status int, message []byte) {
	response.WriteHeader(status)
	_, err := response.Write(message) // TODO подумать, нужно ли
	if err != nil {
		logger.G.Error(err)
	}
}

// GetErrorJSONBody Создание тела ответа с json ошибкой
func GetErrorJSONBody(message string) []byte {
	responseBody := payload.ResponseBody{
		Status:  payload.ResponseErrorStatus,
		Message: message,
	}
	jsonResponse, err := json.Marshal(responseBody)
	if err != nil {
		logger.G.Fatal(err)
	}

	return jsonResponse
}
