package metrics

import (
	"context"
	"io"
)

// Storage represents an interface for accessing and manipulating metrics storage.
type Storage interface {
	// GetGauges получение всех gauge
	GetGauges() (map[string]Gauge, error)
	// GetCounters получение всех counter
	GetCounters() (map[string]Counter, error)
	// SetGauge устанавливаем gauge
	SetGauge(name string, value Gauge) error
	// AddCounter добавляем каунтер
	AddCounter(name string, value Counter) error
	// GetGauge получение отдельного gauge
	GetGauge(name string) (Gauge, bool)
	// GetCounter получение отдельного counter
	GetCounter(name string) (Counter, bool)
	// SetGauges массовое обновление метрик Гауге
	SetGauges(map[string]Gauge) error
	// AddCounters массовое обновление метрик Каунтер
	AddCounters(map[string]Counter) error
}

type SynchronizationStorage interface {
	io.Closer
	// Flush сохраняем не сохранённые элементы
	Flush() error
	// Sync устанавливаем тикер с синхронизацией данных
	Sync(ctx context.Context) error
	// FlushAndClose сохранить несохранённое и закрыть хранилище
	FlushAndClose() error
	// IsSyncMode открыто ли хранилище в синхронном режиме
	IsSyncMode() bool
}

// MeStore Хранилище метрик в памяти.
var MeStore Storage
