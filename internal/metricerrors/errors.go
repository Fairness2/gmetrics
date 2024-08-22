package metricerrors

type Retriable struct {
	original error
}

func (r *Retriable) Error() string {
	return r.original.Error()
}

func (r *Retriable) Unwrap() error {
	return r.original
}

func NewRetriable(err error) error {
	return &Retriable{original: err}
}

var RetriableError = &Retriable{original: nil}
