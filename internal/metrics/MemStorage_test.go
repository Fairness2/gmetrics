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
				if err := store.SetGauges(bm.gauges); err != nil {
					b.Error(err, "error setting gauges")
				}
				if err := store.AddCounters(bm.counters); err != nil {
					b.Error(err, "error add counters")
				}
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

func TestMemStorage_AddCounters(t *testing.T) {
	store := NewMemStorage()
	testCases := []struct {
		name        string
		counters    map[string]Counter
		wantErr     bool
		wantCounter map[string]Counter
	}{
		{
			name: "add_new_counters",
			counters: map[string]Counter{
				"counter1": 1,
				"counter2": 2,
			},
			wantErr: false,
			wantCounter: map[string]Counter{
				"counter1": 1,
				"counter2": 2,
			},
		},
		{
			name: "add_existing_counters",
			counters: map[string]Counter{
				"counter1": 3,
				"counter2": 4,
			},
			wantErr: false,
			wantCounter: map[string]Counter{
				"counter1": 4,
				"counter2": 6,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.AddCounters(tc.counters)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for k, v := range tc.wantCounter {
					assert.Equal(t, v, store.Counter[k])
				}
			}
		})
	}
}

func TestMemStorage_SetGauges(t *testing.T) {
	store := NewMemStorage()
	testCases := []struct {
		name       string
		gauges     map[string]Gauge
		wantErr    bool
		wantGauges map[string]Gauge
	}{
		{
			name: "set_new_gauges",
			gauges: map[string]Gauge{
				"gauge1": 1,
				"gauge2": 2,
			},
			wantErr: false,
			wantGauges: map[string]Gauge{
				"gauge1": 1,
				"gauge2": 2,
			},
		},
		{
			name: "set_existing_counters",
			gauges: map[string]Gauge{
				"gauge1": 3,
				"gauge2": 4,
			},
			wantErr: false,
			wantGauges: map[string]Gauge{
				"gauge1": 3,
				"gauge2": 4,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.SetGauges(tc.gauges)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for k, v := range tc.gauges {
					assert.Equal(t, v, store.Gauge[k])
				}
			}
		})
	}
}

func TestMemStorage_GetGauges(t *testing.T) {
	store := NewMemStorage()
	err := store.SetGauges(map[string]Gauge{
		"gauge1": 0.42,
		"gauge2": 0.84,
	})
	assert.NoError(t, err, "error setting gauges")

	gauges, err := store.GetGauges()
	assert.NoError(t, err, "error getting gauges")
	assert.Len(t, gauges, 2, "expected 2 gauges, got %d", len(gauges))

	g1, ok := gauges["gauge1"]
	assert.True(t, ok, "expected gauge1 to be set")
	assert.Equal(t, g1, Gauge(0.42), "expected gauge1 = 0.42, got %v", g1)

	g2, ok := gauges["gauge2"]
	assert.True(t, ok, "expected gauge1 to be set")
	assert.Equal(t, g2, Gauge(0.84), "expected gauge1 = 0.84, got %v", g1)
}

func TestMemStorage_GetCounters(t *testing.T) {
	store := NewMemStorage()
	err := store.AddCounters(map[string]Counter{
		"counter1": 42,
		"counter2": 84,
	})
	assert.NoError(t, err, "error setting counters")

	counters, err := store.GetCounters()
	assert.NoError(t, err, "error getting counters")
	assert.Len(t, counters, 2, "expected 2 counters, got %d", len(counters))

	c1, ok := counters["counter1"]
	assert.True(t, ok, "expected counter1 to be set")
	assert.Equal(t, c1, Counter(42), "expected counter1 = 42, got %v", c1)

	c2, ok := counters["counter2"]
	assert.True(t, ok, "expected counter2 to be set")
	assert.Equal(t, c2, Counter(84), "expected counter2 = 84, got %v", c1)
}
