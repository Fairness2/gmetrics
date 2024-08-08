package main

import (
	"bytes"
	"compress/gzip"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gmetrics/internal/middlewares"
	store "gmetrics/internal/store/skill"
	"gmetrics/internal/store/skill/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createMockStore(t *testing.T) *mock.MockStore {
	// создадим конроллер моков и экземпляр мок-хранилища
	ctrl := gomock.NewController(t)
	s := mock.NewMockStore(ctrl)

	// определим, какой результат будем получать от «хранилища»
	messages := []store.Message{
		{
			Sender:  "411419e5-f5be-4cdb-83aa-2ca2b6648353",
			Time:    time.Now(),
			Payload: "Hello!",
		},
	}
	// установим условие: при любом вызове метода ListMessages возвращать массив messages без ошибки
	s.EXPECT().
		ListMessages(gomock.Any(), gomock.Any()).
		Return(messages, nil).
		AnyTimes()

	return s
}

func TestWebhook(t *testing.T) {
	// создадим экземпляр приложения и передадим ему «хранилище»
	appInstance := newApp(createMockStore(t))
	// тип http.HandlerFunc реализует интерфейс http.Handler
	// это поможет передать хендлер тестовому серверу
	handler := http.HandlerFunc(appInstance.webhook)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(handler)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		name         string
		method       string
		body         string // добавляем тело запроса в табличные тесты
		expectedCode int
		expectedBody string
	}{
		{
			name:         "sending_using_GET",
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{name: "sending_using_PUT", method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
		{name: "sending_using_DELETE", method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
		{name: "method_post_without_body", method: http.MethodPost, expectedCode: http.StatusInternalServerError, expectedBody: ""},
		{
			name:         "method_post_unsupported_type",
			method:       http.MethodPost,
			expectedCode: http.StatusUnprocessableEntity,
			expectedBody: "",
			body:         `{"request": {"type": "idunno", "command": "do something"}, "version": "1.0"}`,
		},
		{
			name:         "method_post_success",
			method:       http.MethodPost,
			body:         `{"request": {"type": "SimpleUtterance", "command": "sudo do something"}, "session": {"new": true, "user":{"user_id":"6C91DA5198D1758C6A9F63A7C5CDDF09359F683B13A18A151FBF4C8B092BB0C2","access_token":"AgAAAAAB4vpbAAApoR1oaCd5yR6eiXSHqOGT8dT"}}, "version": "1.0"}`,
			expectedCode: http.StatusOK,
			expectedBody: `Точное время .* часов, .* минут. Для вас 1 новых сообщений.`,
		},
		{
			name:         "method_post_success_but_not_messages",
			method:       http.MethodPost,
			body:         `{"request": {"type": "SimpleUtterance", "command": "sudo do something"}, "version": "1.0"}`,
			expectedCode: http.StatusOK,
			expectedBody: `Для вас 1 новых сообщений.`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// делаем запрос с помощью библиотеки resty к адресу запущенного сервера,
			// который хранится в поле URL соответствующей структуры
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL

			if len(tc.body) > 0 {
				req.SetHeader("Content-Type", "application/json")
				req.SetBody(tc.body)
			}

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			// проверяем корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				assert.Regexp(t, tc.expectedBody, string(resp.Body()))
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	// создадим экземпляр приложения и передадим ему «хранилище»
	appInstance := newApp(createMockStore(t))
	handler := Pipeline(http.HandlerFunc(appInstance.webhook), middlewares.GZIPCompressResponse,
		middlewares.GZIPDecompressRequest)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{
        "request": {
            "type": "SimpleUtterance",
            "command": "sudo do something"
        },
        "version": "1.0"
    }`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `{
        "response": {
            "text": "Для вас 1 новых сообщений."
        },
        "version": "1.0"
    }`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err, "error creating gzip writer")
		err = zb.Close()
		require.NoError(t, err, "error closing gzip writer")

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err, "error making HTTP request")
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected HTTP status code")

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "error reading response body")
		require.JSONEq(t, successBody, string(b), "unexpected response body")
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err, "error making HTTP request")
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected HTTP status code")

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err, "error creating gzip reader")

		b, err := io.ReadAll(zr)
		require.NoError(t, err, "error reading response body")

		require.JSONEq(t, successBody, string(b), "unexpected response body")
	})
}
