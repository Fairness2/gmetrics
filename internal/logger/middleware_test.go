package logger

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

type InternalLogger struct {
	logs []string
}

func (l *InternalLogger) Debug(args ...interface{}) {
	l.log("DEBUG", args...)
}

func (l *InternalLogger) Errorf(template string, args ...interface{}) {
	logEntry := fmt.Sprintf("[ERROR] "+template, args...)
	l.logs = append(l.logs, logEntry)
}

func (l *InternalLogger) log(level string, args ...interface{}) {
	logEntry := fmt.Sprintf("[%s] %s", level, fmt.Sprint(args...))
	l.logs = append(l.logs, logEntry)
}

func (l *InternalLogger) Info(args ...interface{}) {
	l.log("INFO", args...)
}

func (l *InternalLogger) Infof(template string, args ...interface{}) {
	logEntry := fmt.Sprintf("[INFO] "+template, args...)
	l.logs = append(l.logs, logEntry)
}

func (l *InternalLogger) Error(args ...interface{}) {
	l.log("ERROR", args...)
}

func (l *InternalLogger) Infow(msg string, keysAndValues ...interface{}) {
	logEntry := fmt.Sprintf("[INFO] %s %v", msg, keysAndValues)
	l.logs = append(l.logs, logEntry)
}

func (l *InternalLogger) Warn(args ...interface{}) {
	l.log("WARN", args...)
}

func (l *InternalLogger) Fatal(args ...interface{}) {
	l.log("FATAL", args...)
}

func (l *InternalLogger) Debugw(msg string, keysAndValues ...interface{}) {
	logEntry := fmt.Sprintf("[DEBUG] %s %v", msg, keysAndValues)
	l.logs = append(l.logs, logEntry)
}

// Optionally, you might want to add a method to retrieve logs for the purpose of testing or other use.
func (l *InternalLogger) GetLogs() []string {
	return l.logs
}

func TestLogRequests(t *testing.T) {
	tests := []struct {
		name            string
		request         func() *resty.Request
		url             string
		wantStatus      int
		wantMessagesCnt int
		wantLogs        []string
	}{
		{
			name:            "get_request",
			wantStatus:      http.StatusOK,
			wantMessagesCnt: 2,
			request: func() *resty.Request {
				request := resty.New().R()
				request.Method = http.MethodGet
				return request
			},
			url:      "/aboba",
			wantLogs: []string{"Got incoming HTTP request [method GET path /aboba]", "Got incoming HTTP request [method GET path /aboba"},
		},
		{
			name:            "post_request",
			wantStatus:      http.StatusOK,
			wantMessagesCnt: 2,
			request: func() *resty.Request {
				request := resty.New().R()
				request.Method = http.MethodPost
				return request
			},
			url:      "/",
			wantLogs: []string{"Got incoming HTTP request [method POST path /]", "Got incoming HTTP request [method POST path /"},
		},
		{
			name:            "post_request_405",
			wantStatus:      http.StatusMethodNotAllowed,
			wantMessagesCnt: 2,
			request: func() *resty.Request {
				request := resty.New().R()
				request.Method = http.MethodPost
				return request
			},
			url:      "/aboba",
			wantLogs: []string{"Got incoming HTTP request [method POST path /]", "Got incoming HTTP request [method POST path /"},
		},
		{
			name:            "post_request_404",
			wantStatus:      http.StatusNotFound,
			wantMessagesCnt: 2,
			request: func() *resty.Request {
				request := resty.New().R()
				request.Method = http.MethodPost
				return request
			},
			url:      "/aboba1",
			wantLogs: []string{"Got incoming HTTP request [method POST path /]", "Got incoming HTTP request [method POST path /"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Log = &InternalLogger{
				logs: make([]string, 2),
			}
			router := chi.NewRouter()
			router.Use(LogRequests)
			router.Post("/", func(writer http.ResponseWriter, request *http.Request) {})
			router.Get("/aboba", func(writer http.ResponseWriter, request *http.Request) {})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			request := tt.request()
			request.URL = srv.URL + tt.url
			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.wantStatus, res.StatusCode())
		})
	}
}
