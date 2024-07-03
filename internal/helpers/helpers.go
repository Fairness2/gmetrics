package helpers

import (
	"fmt"
	"net/http"
)

// SetHTTPError Отправка ошибки и сообщения ошибки.
// Parameters:
// - response: http.ResponseWriter object containing information about the HTTP response
// - status: the HTTP status code to set in the response
// - message: the message to write to the response
func SetHTTPError(response http.ResponseWriter, status int, message string) {
	response.WriteHeader(status)
	fmt.Fprint(response, message)
}
