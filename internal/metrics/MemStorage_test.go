package metrics

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	memStore := NewMemStorage()
	assert.NotNil(t, memStore)
}

func TestSetGauge(t *testing.T) {
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

func TestAddCounter(t *testing.T) {
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

func TestGet(t *testing.T) {
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
