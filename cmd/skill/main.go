// пакеты исполняемых приложений должны называться main
// В уроках в конце каждой главы перед инкрементом даётся пример по пройденному материалу на примере разработки навыка для алисы. Считается отдельным третьим приложением после агента и сервера
package main

import (
	"encoding/json"
	"gmetrics/internal/logger"
	"gmetrics/internal/middlewares"
	models "gmetrics/internal/models/skill"
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
	lgr, err := logger.New(flagLogLevel)
	if err != nil {
		return err
	}
	logger.Log = lgr

	// создаём экземпляр приложения, пока без внешней зависимости хранилища сообщений
	appInstance := newApp(nil)

	logger.Log.Info("Running server", zap.String("address", flagRunAddr))
	// оборачиваем хендлер webhook в middleware с логированием
	return http.ListenAndServe(":8080", Pipeline(http.HandlerFunc(appInstance.webhook),
		logger.LogRequests,
		setHeaders,
		middlewares.GZIPCompressResponse,
		middlewares.GZIPDecompressRequest,
	))
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
