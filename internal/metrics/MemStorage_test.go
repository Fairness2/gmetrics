package metrics

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemStorage(t *testing.T) {
	memStore := NewMemStorage()
	assert.NotNil(t, memStore)
}

func TestMemStorage_SetGauge(t *testing.T) {
	memStore := NewMemStorage()
	testCases := []struct {
		name      string
		mName     string
		wantValue Gauge
	}{
		{
			name:      "set_new_gauge",
			mName:     "metricname",
			wantValue: Gauge(42.5),
		},
		{
			name:      "overwrite_gauge",
			mName:     "metricname",
			wantValue: Gauge(43.5),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := memStore.SetGauge(tc.mName, tc.wantValue)
			require.NoError(t, err)
			v, _ := memStore.GetGauge(tc.mName)
			assert.Equal(t, tc.wantValue, v)
		})
	}
}

func TestMemStorage_UnsafeSetGauge(t *testing.T) {
	memStore := NewMemStorage()
	testCases := []struct {
		name      string
		mName     string
		wantValue Gauge
	}{
		{
			name:      "set_new_gauge",
			mName:     "metricname",
			wantValue: Gauge(42.5),
		},
		{
			name:      "overwrite_gauge",
			mName:     "metricname",
			wantValue: Gauge(43.5),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := memStore.unsafeSetGauge(tc.mName, tc.wantValue)
			require.NoError(t, err)
			v, _ := memStore.GetGauge(tc.mName)
			assert.Equal(t, tc.wantValue, v)
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	memStore := NewMemStorage()
	testCases := []struct {
		name      string
		mName     string
		addValue  Counter
		wantValue Counter
	}{
		{
			name:      "add_new_counter",
			mName:     "metricname",
			addValue:  Counter(5),
			wantValue: Counter(5),
		},
		{
			name:      "increment_counter",
			mName:     "metricname",
			addValue:  Counter(5),
			wantValue: Counter(10),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := memStore.AddCounter(tc.mName, tc.addValue)
			require.NoError(t, err)
			v, _ := memStore.GetCounter(tc.mName)
			assert.Equal(t, tc.wantValue, v)
		})
	}
}

func TestMemStorage_UnsafeAddCounter(t *testing.T) {
	memStore := NewMemStorage()
	testCases := []struct {
		name      string
		mName     string
		addValue  Counter
		wantValue Counter
	}{
		{
			name:      "add_new_counter",
			mName:     "metricname",
			addValue:  Counter(5),
			wantValue: Counter(5),
		},
		{
			name:      "increment_counter",
			mName:     "metricname",
			addValue:  Counter(5),
			wantValue: Counter(10),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := memStore.unsafeAddCounter(tc.mName, tc.addValue)
			require.NoError(t, err)
			v, _ := memStore.GetCounter(tc.mName)
			assert.Equal(t, tc.wantValue, v)
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	memStore := NewMemStorage()
	_ = memStore.SetGauge("temp", 42.5)

	testCases := []struct {
		name      string
		mName     string
		wantValue any
		ok        bool
	}{
		{
			name:      "get_gauge",
			mName:     "temp",
			ok:        true,
			wantValue: Gauge(42.5),
		},
		{
			name:      "get_non-existent_value",
			mName:     "load",
			ok:        false,
			wantValue: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v, ok := memStore.GetGauge(tc.mName)
			if tc.ok {
				assert.True(t, ok)
				assert.Equal(t, tc.wantValue, v)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	memStore := NewMemStorage()
	_ = memStore.AddCounter("hits", 1)

	testCases := []struct {
		name      string
		mName     string
		wantValue any
		ok        bool
	}{
		{
			name:      "get_counter",
			mName:     "hits",
			ok:        true,
			wantValue: Counter(1),
		},
		{
			name:      "get_non-existent_value",
			mName:     "load",
			ok:        false,
			wantValue: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v, ok := memStore.GetCounter(tc.mName)
			if tc.ok {
				assert.True(t, ok)
				assert.Equal(t, tc.wantValue, v)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func BenchmarkMemStorage_SetMetrics(b *testing.B) {
	benchmarks := []struct {
		name     string
		gauges   map[string]Gauge
		counters map[string]Counter
	}{
		{
			name:     "one_value",
			gauges:   generateGaugesMap(1),
			counters: generateCounterMap(1),
		},
		{
			name:     "1000_value",
			gauges:   generateGaugesMap(1000),
			counters: generateCounterMap(1000),
		},
		{
			name:     "10000_value",
			gauges:   generateGaugesMap(1000),
			counters: generateCounterMap(1000),
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			store := NewMemStorage()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store.SetGauges(bm.gauges)
				store.AddCounters(bm.counters)
			}
		})
	}
}

func generateGaugesMap(size int) map[string]Gauge {
	gauges := make(map[string]Gauge, size)
	for i := 0; i < size; i++ {
		gauges[fmt.Sprintf("metric%d", i+1)] = Gauge(rand.Float64() * 100)
	}
	return gauges
}
func generateCounterMap(size int) map[string]Counter {
	gauges := make(map[string]Counter, size)
	for i := 0; i < size; i++ {
		gauges[fmt.Sprintf("metric%d", i+1)] = Counter(rand.Int())
	}
	return gauges
}
