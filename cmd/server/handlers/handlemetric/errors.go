package handlemetric

import (
	"errors"
	"net/http"
)

type UpdateMetricError struct {
	error
	//Message    string
	HTTPStatus int
}

func (e *UpdateMetricError) Error() string {
	return e.error.Error()
}

func (e *UpdateMetricError) Unwrap() error {
	return nil
}

/*func (er UpdateMetricError) Error() string {
	return er.Message
}*/

var NotValidGaugeError = &UpdateMetricError{
	error:      errors.New("metric value is not a valid float"),
	HTTPStatus: http.StatusBadRequest,
}

var NotValidCounterError = &UpdateMetricError{
	error:      errors.New("metric value is not a valid int"),
	HTTPStatus: http.StatusBadRequest,
}

var InvalidMetricTypeError = &UpdateMetricError{
	error:      errors.New("invalid metric type"),
	HTTPStatus: http.StatusBadRequest,
}

var BadRequestError = &UpdateMetricError{
	error:      errors.New("invalid body"),
	HTTPStatus: http.StatusBadRequest,
}
