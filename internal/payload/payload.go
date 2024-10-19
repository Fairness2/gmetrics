package payload

// Metrics описывает структуру данных для представления метрик.
type Metrics struct {
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // Параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge
}

// ResponseSuccessStatus статус, что метрика установлена удачно
var ResponseSuccessStatus = "success"

// ResponseErrorStatus статус, что метрика установлена не удачно
var ResponseErrorStatus = "error"

// ResponseBody представляет структуру типичного тела ответа API.
type ResponseBody struct {
	Status  string  `json:"status"` // Успешный или не успешный результат
	ID      string  `json:"id,omitempty"`
	Message string  `json:"message,omitempty"`
	Delta   int64   `json:"delta,omitempty"` // Новое значение метрики в случае передачи counter
	Value   float64 `json:"value,omitempty"` // Новое значение метрики в случае передачи gauge
}
