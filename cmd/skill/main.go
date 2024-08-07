// пакеты исполняемых приложений должны называться main
// В уроках в конце каждой главы перед инкрементом даётся пример по пройденному материалу на примере разработки навыка для алисы. Считается отдельным третьим приложением после агента и сервера
package main

import (
	"database/sql"
	"gmetrics/internal/logger"
	"gmetrics/internal/middlewares"
	"gmetrics/internal/store/skill/pg"
	"go.uber.org/zap"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	// создаём соединение с СУБД PostgreSQL с помощью аргумента командной строки
	conn, err := sql.Open("pgx", flagDatabaseURI)
	if err != nil {
		return err
	}

	// создаём экземпляр приложения, пока без внешней зависимости хранилища сообщений
	appInstance := newApp(pg.NewStore(conn))

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
