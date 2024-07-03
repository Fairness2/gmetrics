package main

import (
	"github.com/go-chi/chi/v5"
	"gmetrics/cmd/server/handlers/getMetric"
	"gmetrics/cmd/server/handlers/getMetrics"
	"gmetrics/cmd/server/handlers/handleMetric"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8080", getRouter())
	if err != nil {
		panic(err)
	}
}

func getRouter() chi.Router {
	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handleMetric.Handler)
	router.Get("/", getMetrics.Handler)
	router.Get("/value/{type}/{name}", getMetric.Handler)
	return router
}
