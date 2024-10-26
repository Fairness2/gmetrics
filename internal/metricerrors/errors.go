package metricerrors

// Retriable представляет ошибку, которая требует повторной попытки исполнения.
// Оригинал содержит основную ошибку, вызвавшую повторную ошибку.
type Retriable struct {
	original error
}

// Error возвращает сообщение об исходной ошибке.
func (r *Retriable) Error() string {
	return r.original.Error()
}

// Unwrap возвращает исходную ошибку, инкапсулированную в Retriable.
func (r *Retriable) Unwrap() error {
	return r.original
}

// NewRetriable оборачивает ошибку в ошибку Retriable, указывающую, что операцию можно повторить.
func NewRetriable(err error) error {
	return &Retriable{original: err}
}

// RetriableError представляет собой ошибку, требующую повторной попытки, инициализированную без исходной ошибки.
var RetriableError = &Retriable{original: nil}
