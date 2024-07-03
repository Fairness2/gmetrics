package helpers

import (
	"github.com/stretchr/testify/assert"
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
			name:             "OK Status",
			mockStatus:       http.StatusOK,
			mockMessage:      "OK",
			expectedResponse: "OK",
		},
		{
			name:             "Not Found Status",
			mockStatus:       http.StatusNotFound,
			mockMessage:      "Not Found",
			expectedResponse: "Not Found",
		},
		{
			name:             "Internal Server Error Status",
			mockStatus:       http.StatusInternalServerError,
			mockMessage:      "Internal Server Error",
			expectedResponse: "Internal Server Error",
		},
		{
			name:             "Empty Message",
			mockStatus:       http.StatusInternalServerError,
			mockMessage:      "",
			expectedResponse: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			SetHTTPError(response, c.mockStatus, c.mockMessage)
			result := response.Result()

			assert.Equal(t, c.mockStatus, result.StatusCode)

			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)

			assert.NoError(t, err)
			assert.Equal(t, c.expectedResponse, string(body))
		})
	}
}
