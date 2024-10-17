// File: initializer_test.go

package collection

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
