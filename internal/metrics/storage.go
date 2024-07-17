package metrics

// Storage represents an interface for accessing and manipulating metrics storage.
type Storage interface {
	GetGauges() map[string]Gauge
	GetCounters() map[string]Counter
	SetGauge(name string, value Gauge)
	AddCounter(name string, value Counter)
	GetGauge(name string) (Gauge, bool)
	GetCounter(name string) (Counter, bool)
}

// MeStore Хранилище метрик в памяти.
var MeStore Storage

func SetGlobalStorage(storage Storage) {
	MeStore = storage
}
