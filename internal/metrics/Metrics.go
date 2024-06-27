package metrics

// Gauge Тип метрики gauge
type Gauge float64

// Counter Тип метрики counter
type Counter int64

func (c Counter) Add(v Counter) Counter {
	c += v
	return c
}
