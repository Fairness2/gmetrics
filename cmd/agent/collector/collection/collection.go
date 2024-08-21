package collection

import (
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"math/rand/v2"
	"runtime"
	"sync"
)

// Type represents a CollectionType of metrics, including various gauges and a counter.
type Type struct {
	// Values Мапа со значениями собираемых метрик
	Values map[string]any
	// PollCount счётчик, увеличивающийся на 1 при каждом обновлении метрики из пакета runtime
	PollCount metrics.Counter

	// mutex Мьютекс для устранения состояния гонки при параллельных внесениях изменений в кэш
	// mutex TODO Пока убрал RWMutex, так как с обычным работать проще, а в команде на отправку сбрасываем каунтер
	mutex *sync.Mutex
}

// Collect Заполнение коллекции из метрик системы
func (c *Type) Collect(stats runtime.MemStats) {
	logger.Log.Info("Collecting metrics...")
	// Использование мьютекса предотвращает попытки одновременной записи в коллекцию
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Values["Alloc"] = metrics.Gauge(stats.Alloc)
	c.Values["TotalAlloc"] = metrics.Gauge(stats.TotalAlloc)
	c.Values["BuckHashSys"] = metrics.Gauge(stats.BuckHashSys)
	c.Values["Frees"] = metrics.Gauge(stats.Frees)
	c.Values["GCCPUFraction"] = metrics.Gauge(stats.GCCPUFraction)
	c.Values["GCSys"] = metrics.Gauge(stats.GCSys)
	c.Values["HeapAlloc"] = metrics.Gauge(stats.HeapAlloc)
	c.Values["HeapIdle"] = metrics.Gauge(stats.HeapIdle)
	c.Values["HeapInuse"] = metrics.Gauge(stats.HeapInuse)
	c.Values["HeapObjects"] = metrics.Gauge(stats.HeapObjects)
	c.Values["HeapReleased"] = metrics.Gauge(stats.HeapReleased)
	c.Values["HeapSys"] = metrics.Gauge(stats.HeapSys)
	c.Values["LastGC"] = metrics.Gauge(stats.LastGC)
	c.Values["Lookups"] = metrics.Gauge(stats.Lookups)
	c.Values["MCacheInuse"] = metrics.Gauge(stats.MCacheInuse)
	c.Values["MCacheSys"] = metrics.Gauge(stats.MCacheSys)
	c.Values["MSpanInuse"] = metrics.Gauge(stats.MSpanInuse)
	c.Values["MSpanSys"] = metrics.Gauge(stats.MSpanSys)
	c.Values["Mallocs"] = metrics.Gauge(stats.Mallocs)
	c.Values["NextGC"] = metrics.Gauge(stats.NextGC)
	c.Values["NumForcedGC"] = metrics.Gauge(stats.NumForcedGC)
	c.Values["NumGC"] = metrics.Gauge(stats.NumGC)
	c.Values["OtherSys"] = metrics.Gauge(stats.OtherSys)
	c.Values["PauseTotalNs"] = metrics.Gauge(stats.PauseTotalNs)
	c.Values["StackInuse"] = metrics.Gauge(stats.StackInuse)
	c.Values["StackSys"] = metrics.Gauge(stats.StackSys)
	c.Values["Sys"] = metrics.Gauge(stats.Sys)

	c.PollCount = c.PollCount.Add(1)
	c.Values["RandomValue"] = metrics.Gauge(rand.Float64())

	logger.Log.Info("End collecting metrics")
}

// Lock Установка лока на изменение
func (c *Type) Lock() {
	c.mutex.Lock()
}

// Unlock Завершение чтения с collection и освобождение mutex
func (c *Type) Unlock() {
	c.mutex.Unlock()
}

// ResetCounter Обнуление значения счетчика PollCount
func (c *Type) ResetCounter() {
	c.PollCount = c.PollCount.Clear()
}
