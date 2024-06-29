package metrics

import "strconv"

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

// Gauge Тип метрики gauge
type Gauge float64

func (g Gauge) ToString() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

// Counter Тип метрики counter
type Counter int64

func (c Counter) Add(v Counter) Counter {
	c += v
	return c
}

func (c Counter) Clear() Counter {
	return 0
}

func (g Counter) ToString() string {
	return strconv.FormatInt(int64(g), 10)
}
