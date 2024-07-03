package main

import (
	"github.com/go-chi/chi/v5"
	"gmetrics/cmd/server/handlers/getmetric"
	"gmetrics/cmd/server/handlers/getmetrics"
	"gmetrics/cmd/server/handlers/handlemetric"
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
	router.Post("/update/{type}/{name}/{value}", handlemetric.Handler)
	router.Get("/", getmetrics.Handler)
	router.Get("/value/{type}/{name}", getmetric.Handler)
	return router
}
