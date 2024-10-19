package metrics

import (
	"os"
	"testing"
)

func BenchmarkFileStorage_SetMetrics(b *testing.B) {
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
			store, _ := NewFileStorage("bench.json", false, true)
			defer func() {
				_ = store.Close()
				os.Remove("bench.json")
			}()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store.SetGauges(bm.gauges)
				store.AddCounters(bm.counters)
			}
		})
	}
}
