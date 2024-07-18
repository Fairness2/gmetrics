package logger

import (
	"fmt"
	"net/http"
	"time"
)

// LogRequests мидлеваре, которое регистрирует данные запроса
// Функция регистрирует метод, путь и продолжительность каждого запроса
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Регистрируем завершающую функцию, чтобы залогировать в любом случае
		defer func() {
			G.Infow("Got incoming HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start),
			)
		}()
		next.ServeHTTP(w, r)
	})
}

// responseData структура для хранения сведений об ответе
type responseData struct {
	status int
	size   int
}

// responseWriterWithLogging http.ResponseWriter с сохранением метрик ответа для логирования
// содержит в себе responseData, и заполняет её
// композиция с содержанием и расширением http.ResponseWriter
type responseWriterWithLogging struct {
	http.ResponseWriter
	data *responseData
}

// Write реализует метод http.ResponseWriter.Write интерфейса http.ResponseWriter
// Заполняет размер передаваемых данных тела
func (r *responseWriterWithLogging) Write(body []byte) (int, error) {
	size, err := r.ResponseWriter.Write(body)
	r.data.size += size
	return size, err
}

// WriteHeader реализует метод http.ResponseWriter.WriteHeader интерфейса http.ResponseWriter
// Сохраняет статус ответа
func (r *responseWriterWithLogging) WriteHeader(statusCode int) {
	r.data.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// LogResponse мидлеваре, которое регистрирует данные ответа
// Функция регистрирует размер тела и статус ответа
func LogResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		newWriter := &responseWriterWithLogging{
			ResponseWriter: response,
			data:           new(responseData),
		}
		// Регистрируем завершающую функцию, чтобы залогировать в любом случае
		defer func() {
			G.Infow("Sent HTTP response",
				"status", newWriter.data.status,
				"bodySize", fmt.Sprintf("%d B", newWriter.data.size),
			)
		}()
		next.ServeHTTP(newWriter, request)
	})
}
