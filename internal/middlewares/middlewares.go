package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"gmetrics/cmd/server/config"
	"gmetrics/internal/helpers"
	"gmetrics/internal/helpers/compress"
	"gmetrics/internal/logger"
	"io"
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
				helpers.SetHTTPResponse(w, http.StatusInternalServerError, []byte(err.Error()))
				return
			}
			defer func() {
				if cErr := reader.Close(); cErr != nil {
					logger.Log.Warn(cErr)
				}
			}()
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
				defer func() {
					if cErr := writer.Close(); cErr != nil {
						logger.Log.Warn(cErr)
					}
				}()
				writer.Header().Set("Content-Encoding", "gzip")
				newWriter = writer
			}
		}
		next.ServeHTTP(newWriter, r)
	})
}

// CheckSign проверка подписи запроса
func CheckSign(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hashHeader := r.Header.Get("HashSHA256"); hashHeader != "" && config.Params.HashKey != "" {
			hash, err := hex.DecodeString(hashHeader)
			if err != nil {
				helpers.SetHTTPError(w, http.StatusBadRequest, []byte(err.Error()))
				return
			}

			// Читаем тело запроса
			rawBody, err := io.ReadAll(r.Body)
			if err != nil {
				helpers.SetHTTPError(w, http.StatusBadRequest, []byte(err.Error()))
				return
			}
			// Ставим тело снова, чтобы его можно было прочитать снова.
			r.Body = io.NopCloser(bytes.NewBuffer(rawBody))

			harsher := hmac.New(sha256.New, []byte(config.Params.HashKey))
			harsher.Write(rawBody)
			hashSum := harsher.Sum(nil)
			if !hmac.Equal(hash, hashSum) {
				helpers.SetHTTPError(w, http.StatusBadRequest, []byte("body sign is not correct"))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
