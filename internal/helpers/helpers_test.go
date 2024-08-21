package helpers

import (
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
