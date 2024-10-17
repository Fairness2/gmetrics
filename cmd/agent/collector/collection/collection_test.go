package collection

import (
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/metrics"
	"runtime"
	"sync"
	"testing"
)

func TestType_CollectFromMap(t *testing.T) {
	tests := []struct {
		name  string
		stats map[string]metrics.Gauge
	}{
		{
			name:  "empty_metrics",
			stats: map[string]metrics.Gauge{},
		},
		{
			name: "single_metric",
			stats: map[string]metrics.Gauge{
				"example": metrics.Gauge(42.0),
			},
		},
		{
			name: "multiple_metrics",
			stats: map[string]metrics.Gauge{
				"example1": metrics.Gauge(42.0),
				"example2": metrics.Gauge(22),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Type{
				Values: map[string]any{},
				mutex:  &sync.Mutex{},
			}
			c.CollectFromMap(tt.stats)
			assert.Equalf(t, len(tt.stats), len(c.Values), "CollectFromMap() = %v, want %v", len(c.Values), len(tt.stats))

			for name, gauge := range tt.stats {
				val, ok := c.Values[name]
				assert.Truef(t, ok, "CollectFromMap() missing metric name %v", name)
				assert.Equalf(t, gauge, val, "CollectFromMap() = %v, want %v", val, gauge)
			}
		})
	}
}

func TestType_ResetCounter(t *testing.T) {
	tests := []struct {
		name      string
		initValue metrics.Counter
	}{
		{
			name:      "zero_counter",
			initValue: metrics.Counter(0),
		},
		{
			name:      "positive_counter",
			initValue: metrics.Counter(42),
		},
		{
			name:      "large_counter",
			initValue: metrics.Counter(1e9),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Type{
				Values:    map[string]any{},
				PollCount: tt.initValue,
				mutex:     &sync.Mutex{},
			}
			c.ResetCounter()
			assert.Equalf(t, metrics.Counter(0), c.PollCount, "ResetCounter() = %v, want %v", c.PollCount, 0)
		})
	}
}
func TestType_Collect(t *testing.T) {
	tests := []struct {
		name  string
		stats runtime.MemStats
	}{
		{
			name: "minimal_stats",
			stats: runtime.MemStats{
				Alloc:      1,
				TotalAlloc: 1,
			},
		},
		{
			name: "populated_stats",
			stats: runtime.MemStats{
				Alloc:         3000,
				TotalAlloc:    7000,
				BuckHashSys:   259,
				Frees:         92,
				GCCPUFraction: 0.01,
				GCSys:         450,
				HeapAlloc:     2800,
				HeapIdle:      1200,
				HeapInuse:     77,
				HeapObjects:   23,
				HeapReleased:  3800,
				HeapSys:       3021,
				LastGC:        155000,
				Lookups:       500,
				MCacheInuse:   77,
				MCacheSys:     112,
				MSpanInuse:    93,
				MSpanSys:      890,
				Mallocs:       82000,
				NextGC:        10000,
				NumForcedGC:   79,
				NumGC:         7,
				OtherSys:      120,
				PauseTotalNs:  221000,
				StackInuse:    71,
				StackSys:      80,
				Sys:           4000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Type{
				Values: map[string]any{},
				mutex:  &sync.Mutex{},
			}
			c.Collect(tt.stats)
			assert.Len(t, c.Values, len(c.Values))
			for key := range c.Values {
				assert.Contains(t, c.Values, key)
			}
			assert.Equal(t, metrics.Counter(1), c.PollCount)
			assert.GreaterOrEqual(t, c.Values["RandomValue"], 0.0)
			assert.Less(t, c.Values["RandomValue"], 1.0)
		})
	}
}

func BenchmarkType_Collect(b *testing.B) {
	stats := runtime.MemStats{
		Alloc:         3000,
		TotalAlloc:    7000,
		BuckHashSys:   259,
		Frees:         92,
		GCCPUFraction: 0.01,
		GCSys:         450,
		HeapAlloc:     2800,
		HeapIdle:      1200,
		HeapInuse:     77,
		HeapObjects:   23,
		HeapReleased:  3800,
		HeapSys:       3021,
		LastGC:        155000,
		Lookups:       500,
		MCacheInuse:   77,
		MCacheSys:     112,
		MSpanInuse:    93,
		MSpanSys:      890,
		Mallocs:       82000,
		NextGC:        10000,
		NumForcedGC:   79,
		NumGC:         7,
		OtherSys:      120,
		PauseTotalNs:  221000,
		StackInuse:    71,
		StackSys:      80,
		Sys:           4000,
	}
	collection := NewCollection()
	for i := 0; i < b.N; i++ {
		collection.Collect(stats)
	}
}

func BenchmarkType_CollectFromMap(b *testing.B) {
	stats := map[string]metrics.Gauge{
		"example1": metrics.Gauge(42.0),
		"example2": metrics.Gauge(22),
	}
	collection := NewCollection()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collection.CollectFromMap(stats)
	}
}
