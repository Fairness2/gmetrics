package main

import (
	"gmetrics/cmd/server/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.HandleMetric)
	mux.HandleFunc("/metrics/get", handlers.GetMetrics)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
