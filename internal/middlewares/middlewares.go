package middlewares

import (
	"net/http"
)

// JSONHeaders Устанавливаем заголовки свойственные методам с JSON
func JSONHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем необходимые заголовки
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
