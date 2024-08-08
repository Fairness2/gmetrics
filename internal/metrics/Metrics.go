package metrics

import (
	"database/sql/driver"
	"errors"
	"strconv"
)

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

func (g Gauge) Value() (driver.Value, error) {
	return g.GetRaw(), nil
}

func (g *Gauge) Scan(value interface{}) error {
	// если `value` равен `nil`, будет возвращён 0
	if value == nil {
		*g = 0
		return nil
	}

	v, ok := value.(float64)
	if !ok {
		return errors.New("cannot scan value. cannot convert value to float64")
	}
	*g = Gauge(v)

	return nil
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

func (c Counter) Value() (driver.Value, error) {
	return c.GetRaw(), nil
}

func (c *Counter) Scan(value interface{}) error {
	// если `value` равен `nil`, будет возвращён 0
	if value == nil {
		*c = 0
		return nil
	}

	v, ok := value.(int64)
	if !ok {
		return errors.New("cannot scan value. cannot convert value to int64")
	}
	*c = Counter(v)

	return nil
}
