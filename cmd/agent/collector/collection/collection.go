package collection

import (
	"fmt"
	"gmetrics/internal/metrics"
	"math/rand/v2"
	"runtime"
	"sync"
)

// init Инициализация коллекции метрик и связанного мьютекса
func init() {
	mutex = new(sync.RWMutex)
	Collection = newCollection()
}

// Мьютекс для устранения состояния гонки при параллельних внесениях изменений в кэш
var mutex *sync.RWMutex

// Collection represents a CollectionType of metrics, including various gauges and a counter.
var Collection *CollectionType

// CollectionType represents a CollectionType of metrics, including various gauges and a counter.
type CollectionType struct {
	// Alloc is bytes of allocated heap objects.
	Alloc metrics.Gauge
	// TotalAlloc is cumulative bytes allocated for heap objects.
	TotalAlloc metrics.Gauge
	// BuckHashSys is bytes of memory in profiling bucket hash tables.
	BuckHashSys metrics.Gauge
	// Frees is the cumulative count of heap objects freed.
	Frees metrics.Gauge
	// GCCPUFraction is the fraction of this program's available
	// CPU time used by the GC since the program started.
	GCCPUFraction metrics.Gauge
	// GCSys is bytes of memory in garbage collection metadata.
	GCSys metrics.Gauge
	// HeapAlloc is bytes of allocated heap objects.
	HeapAlloc metrics.Gauge
	// HeapIdle is bytes in idle (unused) spans.
	HeapIdle metrics.Gauge
	// HeapInuse is bytes in in-use spans.
	HeapInuse metrics.Gauge
	// HeapObjects is the number of allocated heap objects.
	HeapObjects metrics.Gauge
	// HeapReleased is bytes of physical memory returned to the OS.
	HeapReleased metrics.Gauge
	// HeapSys is bytes of heap memory obtained from the OS.
	HeapSys metrics.Gauge
	// LastGC is the time the last garbage collection finished, as
	// nanoseconds since 1970 (the UNIX epoch).
	LastGC metrics.Gauge
	// Lookups is the number of pointer lookups performed by the
	// runtime.
	//
	// This is primarily useful for debugging runtime internals.
	Lookups metrics.Gauge
	// MCacheInuse is bytes of allocated mcache structures.
	MCacheInuse metrics.Gauge
	// MCacheSys is bytes of memory obtained from the OS for
	// mcache structures.
	MCacheSys metrics.Gauge
	// MSpanInuse is bytes of allocated mspan structures.
	MSpanInuse metrics.Gauge
	// MSpanSys is bytes of memory obtained from the OS for mspan
	// structures.
	MSpanSys metrics.Gauge
	// Mallocs is the cumulative count of heap objects allocated.
	// The number of live objects is Mallocs - Frees.
	Mallocs metrics.Gauge
	// NextGC is the target heap size of the next GC cycle.
	NextGC metrics.Gauge
	// NumForcedGC is the number of GC cycles that were forced by
	// the application calling the GC function.
	NumForcedGC metrics.Gauge
	// NumGC is the number of completed GC cycles.
	NumGC metrics.Gauge
	// OtherSys is bytes of memory in miscellaneous off-heap
	// runtime allocations.
	OtherSys metrics.Gauge
	// PauseTotalNs is the cumulative nanoseconds in GC
	// stop-the-world pauses since the program started.
	PauseTotalNs metrics.Gauge
	// StackInuse is bytes in stack spans.
	StackInuse metrics.Gauge
	// StackSys is bytes of stack memory obtained from the OS.
	StackSys metrics.Gauge
	// Sys is the total bytes of memory obtained from the OS.
	Sys metrics.Gauge

	// PollCount счётчик, увеличивающийся на 1 при каждом обновлении метрики из пакета runtime
	PollCount metrics.Counter
	// RandomValue обновляемое произвольное значение.
	RandomValue metrics.Gauge
}

// newCollection returns a new instance of the CollectionType type, initialized with default values.
func newCollection() *CollectionType {
	return &CollectionType{}
}

// Collect Заполнение коллекции из метрик системы
func (c *CollectionType) Collect(stats runtime.MemStats) {
	fmt.Println("Collecting metrics...")
	// Использование мьютекса предотвращает попытки одновременной записи в коллекцию
	mutex.Lock()
	defer mutex.Unlock()

	c.Alloc = metrics.Gauge(stats.Alloc)
	c.TotalAlloc = metrics.Gauge(stats.TotalAlloc)
	c.BuckHashSys = metrics.Gauge(stats.BuckHashSys)
	c.Frees = metrics.Gauge(stats.Frees)
	c.GCCPUFraction = metrics.Gauge(stats.GCCPUFraction)
	c.GCSys = metrics.Gauge(stats.GCSys)
	c.HeapAlloc = metrics.Gauge(stats.HeapAlloc)
	c.HeapIdle = metrics.Gauge(stats.HeapIdle)
	c.HeapInuse = metrics.Gauge(stats.HeapInuse)
	c.HeapObjects = metrics.Gauge(stats.HeapObjects)
	c.HeapReleased = metrics.Gauge(stats.HeapReleased)
	c.HeapSys = metrics.Gauge(stats.HeapSys)
	c.LastGC = metrics.Gauge(stats.LastGC)
	c.Lookups = metrics.Gauge(stats.Lookups)
	c.MCacheInuse = metrics.Gauge(stats.MCacheInuse)
	c.MCacheSys = metrics.Gauge(stats.MCacheSys)
	c.MSpanInuse = metrics.Gauge(stats.MSpanInuse)
	c.MSpanSys = metrics.Gauge(stats.MSpanSys)
	c.Mallocs = metrics.Gauge(stats.Mallocs)
	c.NextGC = metrics.Gauge(stats.NextGC)
	c.NumForcedGC = metrics.Gauge(stats.NumForcedGC)
	c.NumGC = metrics.Gauge(stats.NumGC)
	c.OtherSys = metrics.Gauge(stats.OtherSys)
	c.PauseTotalNs = metrics.Gauge(stats.PauseTotalNs)
	c.StackInuse = metrics.Gauge(stats.StackInuse)
	c.StackSys = metrics.Gauge(stats.StackSys)
	c.Sys = metrics.Gauge(stats.Sys)

	c.PollCount = c.PollCount.Add(1)
	c.RandomValue = metrics.Gauge(rand.Float64())

	fmt.Println("End collecting metrics")
}

// LockRead Установка лока на изменение
func (c *CollectionType) LockRead() {
	//mutex.RLock()
	mutex.Lock()
}

// UnlockRead Завершение чтения с collection и освобождение mutex
func (c *CollectionType) UnlockRead() {
	//mutex.RUnlock()
	mutex.Unlock()
}

// ResetCounter Обнуление значения счетчика PollCount
func (c *CollectionType) ResetCounter() {
	c.PollCount = c.PollCount.Clear()
}
