// пакеты исполняемых приложений должны называться main
package main

import "net/http"

func main() {
	err := http.ListenAndServe(":8080", Pipeline(http.HandlerFunc(webhook), setHeaders))
	if err != nil {
		panic(err)
	}
}

type Middleware func(next http.Handler) http.Handler

func Pipeline(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}

func setHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		// Устанавливаем разрешённые методы
		response.Header().Set("Access-Control-Allow-Methods", http.MethodPost+", "+http.MethodOptions)
		response.Header().Set("Content-Type", "application/json") // Ставим, что ответ у нас джейсон

		next.ServeHTTP(response, request)
	})
}

func webhook(response http.ResponseWriter, request *http.Request) {
	// Разрешаем только POST запросы
	if request.Method != http.MethodPost {
		response.WriteHeader(http.StatusMethodNotAllowed) // Возвращаем ответ со статусом 405, метод не разрешён
		return
	}
	// Если метод OPTIONS, то отправляем пустой ответ с заголовком с разрешёнными методами
	if request.Method == http.MethodOptions {
		response.WriteHeader(http.StatusNoContent) // Возвращаем ответ со статусом 204, пустой ответ
		return
	}

	_, _ = response.Write([]byte(`{
        "response": {
          "text": "Извините, я пока ничего не умею"
        },
        "version": "1.0"
      }`))

}
