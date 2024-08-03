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

	logger.Log.Info("Running server", zap.String("address", flagRunAddr))
	// оборачиваем хендлер webhook в middleware с логированием
	return http.ListenAndServe(":8080", Pipeline(http.HandlerFunc(webhook),
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

	// десериализуем запрос в структуру модели
	logger.Log.Debug("decoding request")
	var req models.Request

	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// проверяем, что пришёл запрос понятного типа
	if req.Request.Type != models.TypeSimpleUtterance {
		logger.Log.Debug("unsupported request type", zap.String("type", req.Request.Type))
		response.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// заполняем модель ответа
	resp := models.Response{
		Response: models.ResponsePayload{
			Text: "Извините, я пока ничего не умею",
		},
		Version: "1.0",
	}

	response.Header().Set("Content-Type", "application/json")

	// сериализуем ответ сервера
	enc := json.NewEncoder(response)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}
	logger.Log.Debug("sending HTTP 200 response")
}
