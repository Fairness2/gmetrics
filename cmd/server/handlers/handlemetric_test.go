package handlers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseURL(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wantType  string
		wantName  string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "validURL",
			input:     "/update/metrictype/metricname/metricvalue",
			wantType:  "metrictype",
			wantName:  "metricname",
			wantValue: "metricvalue",
			wantErr:   false,
		},
		{
			name:      "incompleteURL",
			input:     "/update/metrictype/metricname/",
			wantType:  "",
			wantName:  "",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "emptySectionURL",
			input:     "/update/metrictype//metricvalue",
			wantType:  "",
			wantName:  "",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "allEmptyURL",
			input:     "/update///",
			wantType:  "",
			wantName:  "",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "moreSectionsURL",
			input:     "/update/metrictype/metricname/metricvalue/extra",
			wantType:  "",
			wantName:  "",
			wantValue: "",
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotType, gotName, gotValue, err := parseURL(tc.input)
			if tc.wantErr {
				assert.Error(t, err, "Not error parsing URL")
			} else {
				assert.NoError(t, err, "Unexpected error parsing URL")
				assert.Equal(t, tc.wantType, gotType)
				assert.Equal(t, tc.wantName, gotName)
				assert.Equal(t, tc.wantValue, gotValue)
			}
		})
	}
}

func TestHandleMetric(t *testing.T) {
	// urlTemplate Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var urlUpdateTemplate = "/update/%s/%s/%s"
	tests := []struct {
		name            string
		sendURL         string
		wantStatus      int
		wantContentType string
	}{
		{
			name:            "Wrong URL",
			sendURL:         "/update",
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "Empty Type",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, "", "someName", "123"),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "Empty Name",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "", "123"),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "Empty Value",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "someName", ""),
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
		},
		{
			name:            "Wrong type",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, "aboba", "someName", "123"),
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "Wrong value gauge",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "someName", "some"),
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "Wrong value count",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeCounter, "someName", "some"),
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
		},
		{
			name:            "Right value gauge",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeGauge, "someName", "56.78"),
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
		{
			name:            "Right value count",
			sendURL:         fmt.Sprintf(urlUpdateTemplate, metrics.TypeCounter, "someName", "5"),
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.sendURL, nil)
			w := httptest.NewRecorder()
			HandleMetric(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.wantStatus, res.StatusCode)
		})
	}
}
