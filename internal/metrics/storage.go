package metrics

// Storage represents an interface for accessing and manipulating metrics storage.
type Storage interface {
	GetGauges() (map[string]Gauge, error)
	GetCounters() (map[string]Counter, error)
	SetGauge(name string, value Gauge) error
	AddCounter(name string, value Counter) error
	GetGauge(name string) (Gauge, bool)
	GetCounter(name string) (Counter, bool)
}

// MeStore Хранилище метрик в памяти.
var MeStore Storage
