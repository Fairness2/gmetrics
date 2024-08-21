package helpers

import (
	"encoding/json"
	"gmetrics/internal/logger"
	"gmetrics/internal/payload"
	"net/http"
)

// SetHTTPResponse Отправка ошибки и сообщения ошибки.
// Parameters:
// - response: http.ResponseWriter object containing information about the HTTP response
// - status: the HTTP status code to set in the response
// - message: the message to write to the response
func SetHTTPResponse(response http.ResponseWriter, status int, message []byte) {
	response.WriteHeader(status)
	_, err := response.Write(message) // TODO подумать, нужно ли
	if err != nil {
		logger.Log.Error(err)
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
		logger.Log.Fatal(err)
	}

	return jsonResponse
}
