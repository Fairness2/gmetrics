// пакеты исполняемых приложений должны называться main
package main

import (
	logger "gmetrics/internal/logger/skill"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	parseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}
func run() error {
	// Создаём логер
	if err := logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	logger.Log.Info("Running server", zap.String("address", flagRunAddr))
	// оборачиваем хендлер webhook в middleware с логированием
	return http.ListenAndServe(":8080", Pipeline(http.HandlerFunc(webhook), logger.RequestLogger, setHeaders))
}

type Middleware func(next http.Handler) http.Handler

func Pipeline(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}

func setHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		// Устанавливаем разрешённые методы
		response.Header().Set("Access-Control-Allow-Methods", http.MethodPost+", "+http.MethodOptions)
		response.Header().Set("Content-Type", "application/json") // Ставим, что ответ у нас джейсон

		next.ServeHTTP(response, request)
	})
}

func webhook(response http.ResponseWriter, request *http.Request) {
	// Разрешаем только POST запросы
	if request.Method != http.MethodPost {
		logger.Log.Debug("got request with bad method", zap.String("method", request.Method))
		response.WriteHeader(http.StatusMethodNotAllowed) // Возвращаем ответ со статусом 405, метод не разрешён
		return
	}
	// Если метод OPTIONS, то отправляем пустой ответ с заголовком с разрешёнными методами
	if request.Method == http.MethodOptions {
		response.WriteHeader(http.StatusNoContent) // Возвращаем ответ со статусом 204, пустой ответ
		return
	}

	_, _ = response.Write([]byte(`{
        "response": {
          "text": "Извините, я пока ничего не умею"
        },
        "version": "1.0"
      }`))
	logger.Log.Debug("sending HTTP 200 response")
}
