package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"gmetrics/cmd/server/config"
	"gmetrics/cmd/server/handlers/getmetric"
	"gmetrics/cmd/server/handlers/getmetrics"
	"gmetrics/cmd/server/handlers/handlemetric"
	"net/http"
)

func main() {
	config.Parse()                // Заполняем конфигурацию сервера
	if err := run(); err != nil { // Запускаем сервер
		panic(err)
	}
}

// run запуск сервера
func run() error {
	fmt.Println("Running server on", config.Params.Address)
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
