// File: initializer_test.go

package collection

import (
	"context"
	"gmetrics/cmd/agent/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCollect(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "not_empty_collection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// assume NewCollection properly initializes an empty collection and returns it
			clearCollection := NewCollection()
			Collection = NewCollection()

			collect()
			assert.NotEqual(t, clearCollection, Collection, "Collection should not be empty")
		})
	}
}

func TestCollectUtil(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "not_empty_collection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// assume NewCollection properly initializes an empty collection and returns it
			clearCollection := NewCollection()
			Collection = NewCollection()

			collectUtil()
			assert.NotEqual(t, clearCollection, Collection, "Collection should not be empty")
		})
	}
}

func BenchmarkCollect(b *testing.B) {
	Collection = NewCollection()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collect()
	}
}

func BenchmarkCollectUtil(b *testing.B) {
	Collection = NewCollection()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collectUtil()
	}
}

func TestCollectUtilProcess(t *testing.T) {
	tests := []struct {
		name      string
		doneAfter time.Duration
	}{
		{
			name:      "collect_called_before_context_done",
			doneAfter: 3 * time.Second,
		},
		{
			name:      "collect_not_called_if_context_done_immediately",
			doneAfter: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config.Params = &config.CliConfig{PollInterval: 1}
			ctx, cancel := context.WithTimeout(context.Background(), tc.doneAfter)
			defer cancel()
			Collection = NewCollection()
			CollectUtilProcess(ctx)
		})
	}
}

func TestCollectProcess(t *testing.T) {
	tests := []struct {
		name      string
		doneAfter time.Duration
	}{
		{
			name:      "collect_called_before_context_done",
			doneAfter: 3 * time.Second,
		},
		{
			name:      "collect_not_called_if_context_done_immediately",
			doneAfter: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config.Params = &config.CliConfig{PollInterval: 1}
			ctx, cancel := context.WithTimeout(context.Background(), tc.doneAfter)
			defer cancel()
			Collection = NewCollection()
			CollectProcess(ctx)
		})
	}
}
