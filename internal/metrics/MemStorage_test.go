package metrics

import (
	"github.com/stretchr/testify/assert"
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
			name:      "Set new Gauge",
			mName:     "metricname",
			wantValue: Gauge(42.5),
		},
		{
			name:      "Overwrite Gauge",
			mName:     "metricname",
			wantValue: Gauge(43.5),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStore.SetGauge(tc.mName, tc.wantValue)
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
			name:      "Add new Counter",
			mName:     "metricname",
			addValue:  Counter(5),
			wantValue: Counter(5),
		},
		{
			name:      "Increment Counter",
			mName:     "metricname",
			addValue:  Counter(5),
			wantValue: Counter(10),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStore.AddCounter(tc.mName, tc.addValue)
			v, _ := memStore.GetCounter(tc.mName)
			assert.Equal(t, tc.wantValue, v)
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	memStore := NewMemStorage()
	memStore.SetGauge("temp", 42.5)

	testCases := []struct {
		name      string
		mName     string
		wantValue interface{}
		ok        bool
	}{
		{
			name:      "Get Gauge",
			mName:     "temp",
			ok:        true,
			wantValue: Gauge(42.5),
		},
		{
			name:      "Get non-existent value",
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
	memStore.AddCounter("hits", 1)

	testCases := []struct {
		name      string
		mName     string
		wantValue interface{}
		ok        bool
	}{
		{
			name:      "Get Counter",
			mName:     "hits",
			ok:        true,
			wantValue: Counter(1),
		},
		{
			name:      "Get non-existent value",
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
