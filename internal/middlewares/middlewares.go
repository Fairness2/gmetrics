package middlewares

import (
	"gmetrics/internal/helpers"
	"gmetrics/internal/helpers/compress"
	"gmetrics/internal/logger"
	"net/http"
	"strings"
)

// JSONHeaders Устанавливаем заголовки свойственные методам с JSON
func JSONHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем необходимые заголовки
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// GZIPDecompressRequest расшифровка сжатого тела запроса в формате gzip
func GZIPDecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если указано сжатие тела в gzip, то заменяем тело на разжатое
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := compress.NewGZIPReader(r.Body)
			logger.Log.Debugw("Content encoded", "type", "gzip")
			if err != nil {
				// Если ошибка создания читателя, то отправляем ошибку сервера
				logger.Log.Error(err)
				helpers.SetHTTPError(w, http.StatusInternalServerError, []byte(err.Error()))
				return
			}
			defer reader.Close()
			r.Body = reader
		}
		next.ServeHTTP(w, r)
	})
}

// GZIPCompressResponse сжатие тела в формат gzip
func GZIPCompressResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newWriter := w
		// Если доступно сжатие, то заменяем писателя на сжимающего и проставляем заголовок, что тело сжато
		if header := r.Header.Get("Accept-Encoding"); strings.Contains(header, "gzip") {
			logger.Log.Debugw("Allowed content encoding", "type", "gzip")
			writer, err := compress.NewGZIPHTTPWriter(w)
			if err != nil {
				logger.Log.Error(err)
			} else {
				defer writer.Close()
				writer.Header().Set("Content-Encoding", "gzip")
				newWriter = writer
			}
		}
		next.ServeHTTP(newWriter, r)
	})
}
