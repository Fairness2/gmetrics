package handlemetric

import (
	"errors"
	"net/http"
)

// UpdateMetricError представляет ошибки, возникшие при обновлении метрики.
// Он включает тип ошибки и поле HTTPStatus для кодов ответов HTTP.
type UpdateMetricError struct {
	error
	//Message    string
	HTTPStatus int
}

// Error возвращает сообщение об ошибке UpdateMetricError.
func (e *UpdateMetricError) Error() string {
	return e.error.Error()
}

// Unwrap возвращает внутреннюю ошибку, вложенную в UpdateMetricError, или nil, если такой нет.
func (e *UpdateMetricError) Unwrap() error {
	return nil
}

/*func (er UpdateMetricError) Error() string {
	return er.Message
}*/

// NotValidGaugeError возвращается, когда значение метрики невозможно преобразовать в допустимый тип Gauge
var NotValidGaugeError = &UpdateMetricError{
	error:      errors.New("metric value is not a valid float"),
	HTTPStatus: http.StatusBadRequest,
}

// NotValidCounterError представляет ошибку, когда значение метрики не может быть преобразовано в Counter
var NotValidCounterError = &UpdateMetricError{
	error:      errors.New("metric value is not a valid int"),
	HTTPStatus: http.StatusBadRequest,
}

// InvalidMetricTypeError представляет ошибку, если указанный тип метрики недействителен.
var InvalidMetricTypeError = &UpdateMetricError{
	error:      errors.New("invalid metric type"),
	HTTPStatus: http.StatusBadRequest,
}

// BadRequestError представляет ошибку, когда тело запроса недействительно.
var BadRequestError = &UpdateMetricError{
	error:      errors.New("invalid body"),
	HTTPStatus: http.StatusBadRequest,
}
