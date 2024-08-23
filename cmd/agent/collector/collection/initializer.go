package collection

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"gmetrics/cmd/agent/config"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
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

// CollectProcess continuously collects system metrics by reading memory stats,
// collecting the stats using collection.Collection.Collect, and then sleeping
// for the duration defined by config.PollInterval.
func CollectProcess(ctx context.Context) {
	logger.Log.Infof("Collect metrics process starts. Period is %d seconds\n", config.Params.PollInterval)
	ticker := time.NewTicker(time.Duration(config.Params.PollInterval) * time.Second)
	// Делаем первый сбор метрик сразу же
	collect()
	for {
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-ticker.C:
			collect()
		case <-ctx.Done():
			ticker.Stop()
			logger.Log.Info("Collect metrics process stopped")
			return
		}
	}
}

// collect reads memory stats and calls Collection.Collect to populate the collection with system metrics.
func collect() {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)
	Collection.Collect(stats)
}

// CollectUtilProcess процесс сбора метрик использования системы
func CollectUtilProcess(ctx context.Context) {
	logger.Log.Infof("Collect util metrics process starts. Period is %d seconds\n", config.Params.PollInterval)
	ticker := time.NewTicker(time.Duration(config.Params.PollInterval) * time.Second)
	// Делаем первый сбор метрик сразу же
	collectUtil()
	for {
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-ticker.C:
			collectUtil()
		case <-ctx.Done():
			ticker.Stop()
			logger.Log.Info("Collect metrics process stopped")
			return
		}
	}
}

// collectUtil собираем коллекцию использования системы
func collectUtil() {
	statsMap := make(map[string]metrics.Gauge)
	memStat, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error(err)
	}
	statsMap["TotalMemory"] = metrics.Gauge(memStat.Total)
	statsMap["FreeMemory"] = metrics.Gauge(memStat.Free)

	cpuStat, err := cpu.Percent(0, true)
	if err != nil {
		logger.Log.Error(err)
	}
	for i, percent := range cpuStat {
		statsMap[fmt.Sprintf("CPUutilization%d", i)] = metrics.Gauge(percent)
	}

	Collection.CollectFromMap(statsMap)
}
