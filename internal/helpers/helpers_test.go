package helpers

import (
	"gmetrics/internal/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetHTTPError(t *testing.T) {
	cases := []struct {
		name             string
		mockStatus       int
		mockMessage      string
		expectedResponse string
	}{
		{
			name:             "OK_status",
			mockStatus:       http.StatusOK,
			mockMessage:      "OK",
			expectedResponse: "OK",
		},
		{
			name:             "not_found_status",
			mockStatus:       http.StatusNotFound,
			mockMessage:      "Not Found",
			expectedResponse: "Not Found",
		},
		{
			name:             "internal_server_error_status",
			mockStatus:       http.StatusInternalServerError,
			mockMessage:      "Internal Server Error",
			expectedResponse: "Internal Server Error",
		},
		{
			name:             "empty_message",
			mockStatus:       http.StatusInternalServerError,
			mockMessage:      "",
			expectedResponse: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			SetHTTPResponse(response, c.mockStatus, []byte(c.mockMessage))
			result := response.Result()
			defer func() {
				if cErr := result.Body.Close(); cErr != nil {
					logger.Log.Warn(cErr)
				}
			}()

			assert.Equal(t, c.mockStatus, result.StatusCode)
			body, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			assert.Equal(t, c.expectedResponse, string(body))
		})
	}
}

func TestGetErrorJSONBody(t *testing.T) {
	cases := []struct {
		name         string
		message      string
		expectedData string
	}{
		{
			name:         "standard_error",
			message:      "An error occurred",
			expectedData: "{\"status\":\"error\",\"message\":\"An error occurred\"}",
		},
		{
			name:         "empty_error",
			message:      "",
			expectedData: "{\"status\":\"error\"}",
		},
		{
			name:         "long_error",
			message:      strings.Repeat("A", 1024),
			expectedData: "{\"status\":\"error\",\"message\":\"" + strings.Repeat("A", 1024) + "\"}",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := GetErrorJSONBody(c.message)
			assert.Equal(t, c.expectedData, string(result))
		})
	}
}
