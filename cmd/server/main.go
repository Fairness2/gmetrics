package main

import (
	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	"gmetrics/cmd/server/config"
	"gmetrics/cmd/server/handlers/getmetric"
	"gmetrics/cmd/server/handlers/getmetrics"
	"gmetrics/cmd/server/handlers/handlemetric"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"gmetrics/internal/middlewares"
	"log"
	"net/http"
)

func main() {
	// Устанавливаем настройки
	cnf, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	config.Params = cnf
	// Инициализируем логер
	lgr, err := logger.New(cnf.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.G = lgr
	// Показываем конфигурацию сервера
	logger.G.Infow("Running server with configuration",
		"address", cnf.Address,
		"logLevel", cnf.LogLevel,
	)
	// Устанавливаем глобальное хранилище метрик
	metrics.MeStore = metrics.NewMemStorage()

	if err := run(); err != nil { // Запускаем сервер
		log.Fatal(err)
	}
}

// run запуск сервера
func run() error {
	logger.G.Infof("Running server on %s", config.Params.Address)
	return http.ListenAndServe(config.Params.Address, getRouter())
}

// getRouter конфигурация роутинга приложение
func getRouter() chi.Router {
	router := chi.NewRouter()
	// Устанавилваем мидлваре с логированием запросов
	router.Use(cMiddleware.StripSlashes, logger.LogResponse, logger.LogRequests)
	// Сохранение метрики по URL
	router.Post("/update/{type}/{name}/{value}", handlemetric.URLHandler)
	// Получение всех метрик
	router.Get("/", getmetrics.Handler)
	// Получение отдельной метрики
	router.Get("/value/{type}/{name}", getmetric.URLHandler)

	router.Group(func(r chi.Router) {
		// Устанавилваем мидлваре с логированием запросов
		r.Use(middlewares.JSONHeaders)
		// Сохранение метрики с помощью JSON тела
		r.Post("/update", handlemetric.JSONHandler)
		// Получение отдельной метрики
		r.Post("/value", getmetric.JSONHandler)
	})
	return router
}
