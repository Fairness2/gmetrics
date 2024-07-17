package collection

import (
	"context"
	"gmetrics/cmd/agent/config"
	"log"
	"runtime"
	"sync"
	"time"
)

// Collection represents a CollectionType of metrics, including various gauges and a counter.
// Collection TODO уйти от глобальных переменных
var Collection *Type

// NewCollection returns a new instance of the CollectionType type, initialized with default values.
func NewCollection() *Type {
	c := Type{
		Values: make(map[string]any),
		mutex:  new(sync.Mutex),
	}
	return &c
}

// SetGlobalCollection Установка глобальной коллекции
func SetGlobalCollection(c *Type) {
	Collection = c
}

// CollectProcess continuously collects system metrics by reading memory stats,
// collecting the stats using collection.Collection.Collect, and then sleeping
// for the duration defined by config.PollInterval.
func CollectProcess(ctx context.Context) {
	log.Printf("Collect metrics process starts. Period is %d mseconds\n", config.Params.PollInterval)
	for {
		stats := runtime.MemStats{}
		runtime.ReadMemStats(&stats)
		Collection.Collect(stats)
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-ctx.Done():
			log.Println("Collect metrics process stopped")
			return
		default:
			//continue
		}
		time.Sleep(config.Params.PollInterval)
	}
}
