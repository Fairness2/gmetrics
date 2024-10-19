package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewFileStorage(t *testing.T) {
	defer os.Remove("test.json")
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	assert.NotNil(t, memStore)
}

func TestFileStorage_SetGauge(t *testing.T) {
	defer os.Remove("test.json")
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
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

func TestFileStorage_AddCounter(t *testing.T) {
	defer os.Remove("test.json")
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
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

func TestFileStorage_GetGauge(t *testing.T) {
	defer os.Remove("test.json")
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
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

func TestFileStorage_GetCounter(t *testing.T) {
	defer os.Remove("test.json")
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
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

func TestRestoreFromFile(t *testing.T) {
	testCases := []struct {
		name        string
		fileName    string
		isErr       bool
		hasFile     bool
		fileContent string
	}{
		{
			name:        "valid_file",
			fileName:    "test.json",
			isErr:       false,
			hasFile:     true,
			fileContent: `{"gauge":{"Alloc":3009592},"counter":{"GetSet75":2210657517}}`,
		},
		{
			name:        "invalid_file",
			fileName:    "non_existing.json",
			isErr:       true,
			hasFile:     true,
			fileContent: `{"gauge":{"Alloc":3009592},"counter":{"GetSet75":2210657517},}`,
		},
		{
			name:     "no_file",
			fileName: "non_existing.json",
			isErr:    true,
			hasFile:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer os.Remove(tc.fileName)
			if tc.hasFile {
				_ = os.WriteFile(tc.fileName, []byte(tc.fileContent), 0644)
			}
			memStore := DurationFileStorage{Storage: NewMemStorage()}
			err := restoreFromFile(tc.fileName, memStore.Storage)
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
