package main

import (
	"github.com/go-chi/chi/v5"
	"gmetrics/cmd/server/config"
	"gmetrics/cmd/server/handlers/getmetric"
	"gmetrics/cmd/server/handlers/getmetrics"
	"gmetrics/cmd/server/handlers/handlemetric"
	"gmetrics/internal/metrics"
	"log"
	"net/http"
)

func main() {
	// Устанавливаем настройки
	cnf, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	config.SetGlobalConfig(cnf)
	log.Print(config.PrintConfig(cnf))

	// Устанавливаем глобальное хранилище метрик
	metrics.MeStore = metrics.NewMemStorage()

	if err := run(); err != nil { // Запускаем сервер
		log.Fatal(err)
	}
}

// run запуск сервера
func run() error {
	log.Println("Running server on", config.Params.Address)
	return http.ListenAndServe(config.Params.Address, getRouter())
}

// getRouter конфигурация роутинга приложение
func getRouter() chi.Router {
	router := chi.NewRouter()
	// Сохранение метрики
	router.Post("/update/{type}/{name}/{value}", handlemetric.Handler)
	// Получение всех метрик
	router.Get("/", getmetrics.Handler)
	// Получение отдельной метрики
	router.Get("/value/{type}/{name}", getmetric.Handler)
	return router
}
