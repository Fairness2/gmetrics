package metrics

import (
	"database/sql/driver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGauge_ToString(t *testing.T) {
	tests := []struct {
		name   string
		gauge  Gauge
		expect string
	}{
		{
			name:   "zero_value",
			gauge:  0,
			expect: "0",
		},
		{
			name:   "positive_value",
			gauge:  7.345,
			expect: "7.345",
		},
		{
			name:   "negative_value",
			gauge:  -6.789,
			expect: "-6.789",
		},
		{
			name:   "large_value",
			gauge:  987654321.1234567,
			expect: "987654321.1234567",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.gauge.ToString())
		})
	}
}

func TestGauge_GetRaw(t *testing.T) {
	tests := []struct {
		name   string
		gauge  Gauge
		expect float64
	}{
		{
			name:   "zero_value",
			gauge:  0,
			expect: 0,
		},
		{
			name:   "positive_value",
			gauge:  7.345,
			expect: 7.345,
		},
		{
			name:   "negative_value",
			gauge:  -6.789,
			expect: -6.789,
		},
		{
			name:   "large_value",
			gauge:  987654321.1234567,
			expect: 987654321.1234567,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.gauge.GetRaw())
		})
	}
}

func TestGauge_Value(t *testing.T) {
	tests := []struct {
		name   string
		gauge  Gauge
		expect driver.Value
	}{
		{
			name:   "zero_value",
			gauge:  0,
			expect: driver.Value(float64(0)),
		},
		{
			name:   "positive_value",
			gauge:  7.345,
			expect: driver.Value(7.345),
		},
		{
			name:   "negative_value",
			gauge:  -6.789,
			expect: driver.Value(-6.789),
		},
		{
			name:   "large_value",
			gauge:  987654321.1234567,
			expect: driver.Value(987654321.1234567),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.gauge.Value()
			assert.NoError(t, err)
			assert.Equal(t, test.expect, value)
		})
	}
}

func TestGauge_Scan(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		expect    Gauge
		expectErr bool
	}{
		{
			name:   "zero_value",
			value:  float64(0),
			expect: 0,
		},
		{
			name:   "positive_value",
			value:  7.345,
			expect: 7.345,
		},
		{
			name:   "negative_value",
			value:  -6.789,
			expect: -6.789,
		},
		{
			name:   "large_value",
			value:  987654321.1234567,
			expect: 987654321.1234567,
		},
		{
			name:   "nil_value",
			value:  nil,
			expect: 0,
		},
		{
			name:      "invalid_value",
			value:     "invalid",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var g Gauge
			err := g.Scan(test.value)
			if test.expectErr {
				assert.Error(t, err)
				assert.ErrorIs(t, ErrorScanGauge, err, "got error is not of type ErrorScanGauge")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expect, g)
			}
		})
	}
}

func TestCounter_ToString(t *testing.T) {
	tests := []struct {
		name    string
		counter Counter
		expect  string
	}{
		{
			name:    "zero_value",
			counter: 0,
			expect:  "0",
		},
		{
			name:    "positive_value",
			counter: 7,
			expect:  "7",
		},
		{
			name:    "negative_value",
			counter: -6,
			expect:  "-6",
		},
		{
			name:    "large_value",
			counter: 987654321,
			expect:  "987654321",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.counter.ToString())
		})
	}
}

func TestCounter_GetRaw(t *testing.T) {
	tests := []struct {
		name    string
		counter Counter
		expect  int64
	}{
		{
			name:    "zero_value",
			counter: 0,
			expect:  0,
		},
		{
			name:    "positive_value",
			counter: 7,
			expect:  7,
		},
		{
			name:    "negative_value",
			counter: -6,
			expect:  -6,
		},
		{
			name:    "large_value",
			counter: 987654321,
			expect:  987654321,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.counter.GetRaw())
		})
	}
}

func TestCounter_Value(t *testing.T) {
	tests := []struct {
		name    string
		counter Counter
		expect  driver.Value
	}{
		{
			name:    "zero_value",
			counter: 0,
			expect:  driver.Value(int64(0)),
		},
		{
			name:    "positive_value",
			counter: 7,
			expect:  driver.Value(int64(7)),
		},
		{
			name:    "negative_value",
			counter: -6,
			expect:  driver.Value(int64(-6)),
		},
		{
			name:    "large_value",
			counter: 987654321,
			expect:  driver.Value(int64(987654321)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.counter.Value()
			assert.NoError(t, err)
			assert.Equal(t, test.expect, value)
		})
	}
}

func TestCounter_Scan(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		expect    Counter
		expectErr bool
	}{
		{
			name:   "zero_value",
			value:  int64(0),
			expect: 0,
		},
		{
			name:   "positive_value",
			value:  int64(7),
			expect: 7,
		},
		{
			name:   "negative_value",
			value:  int64(-6),
			expect: -6,
		},
		{
			name:   "large_value",
			value:  int64(987654321),
			expect: 987654321,
		},
		{
			name:   "nil_value",
			value:  nil,
			expect: 0,
		},
		{
			name:      "invalid_value",
			value:     "invalid",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var c Counter
			err := c.Scan(test.value)
			if test.expectErr {
				assert.Error(t, err)
				assert.ErrorIs(t, ErrorScanCounter, err, "got error is not of type ErrorScanCounter")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expect, c)
			}
		})
	}
}
func TestCounter_Add(t *testing.T) {
	tests := []struct {
		name    string
		initial Counter
		add     Counter
		expect  Counter
	}{
		{
			name:    "zero_add_zero",
			initial: 0,
			add:     0,
			expect:  0,
		},
		{
			name:    "zero_add_positive",
			initial: 0,
			add:     7,
			expect:  7,
		},
		{
			name:    "zero_add_negative",
			initial: 0,
			add:     -6,
			expect:  -6,
		},
		{
			name:    "positive_add_positive",
			initial: 5,
			add:     6,
			expect:  11,
		},
		{
			name:    "positive_add_negative",
			initial: 6,
			add:     -4,
			expect:  2,
		},
		{
			name:    "negative_add_negative",
			initial: -5,
			add:     -3,
			expect:  -8,
		},
		{
			name:    "negative_add_positive",
			initial: -5,
			add:     3,
			expect:  -2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.initial.Add(test.add)
			assert.Equal(t, test.expect, result)
		})
	}
}

func TestCounter_Clear(t *testing.T) {
	tests := []struct {
		name    string
		initial Counter
		expect  Counter
	}{
		{
			name:    "zero_value",
			initial: 0,
			expect:  0,
		},
		{
			name:    "positive_value",
			initial: 7,
			expect:  0,
		},
		{
			name:    "negative_value",
			initial: -6,
			expect:  0,
		},
		{
			name:    "large_value",
			initial: 987654321,
			expect:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.initial.Clear()
			assert.Equal(t, test.expect, result)
		})
	}
}
