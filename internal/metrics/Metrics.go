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

func (g Gauge) GetRaw() float64 {
	return float64(g)
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

func (c Counter) ToString() string {
	return strconv.FormatInt(int64(c), 10)
}

func (c Counter) GetRaw() int64 {
	return int64(c)
}
