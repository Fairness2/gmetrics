package logger

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

// LogRequests мидлеваре, которое регистрирует данные запроса
// Функция регистрирует метод, путь и продолжительность каждого запроса
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		start := time.Now()
		newWriter := &responseWriterWithLogging{
			ResponseWriter: response,
			data:           new(responseData),
		}
		Log.Infow("Got incoming HTTP request",
			"method", request.Method,
			"path", request.URL.Path,
		)
		// Регистрируем завершающую функцию, чтобы залогировать в любом случае
		defer func() {
			Log.Infow("Got incoming HTTP request",
				"method", request.Method,
				"path", request.URL.Path,
				"duration", time.Since(start),
				"status", newWriter.data.status,
				"bodySize", fmt.Sprintf("%d B", newWriter.data.size),
			)
		}()
		next.ServeHTTP(newWriter, request)
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

// LogInterceptor интерсептор, который регистрирует данные запроса по rpc
// Функция регистрирует метод, путь и продолжительность каждого запроса
func LogInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	start := time.Now()
	Log.Infow("Got incoming RPC request",
		"method", "RPC",
		"path", info.FullMethod,
	)
	// Регистрируем завершающую функцию, чтобы залогировать в любом случае
	defer func() {
		Log.Infow("Got incoming RPC request",
			"method", "RPC",
			"path", info.FullMethod,
			"duration", time.Since(start),
			"status", status.Code(err),
		)
	}()
	resp, err = handler(ctx, req)
	return resp, err
}
