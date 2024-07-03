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

// MemStorage Хранилище метрик в памяти
type MemStorage struct {
	gauge   map[string]Gauge
	counter map[string]Counter
	//metrics map[string]interface{}
}

// Set Установка значения метрики в хранилище
// Parameters:
//
//	name: the name of the metric
//	value: the value to be stored for the metric
/*func (storage *MemStorage) Set(name string, value interface{}) {
	storage.metrics[name] = value
}*/

func (storage *MemStorage) SetGauge(name string, value Gauge) {
	storage.gauge[name] = value
}

func (storage *MemStorage) AddCounter(name string, value Counter) {
	oldValue, ok := storage.counter[name]
	if ok {
		value = oldValue.Add(value)
	}
	storage.counter[name] = value
}

// Get Получение значения метрики из хранилища
// Parameters:
//
//	name: имя метрики
//
// Returns:
//
//	value: значение метрики
//	ok: флаг, указывающий на наличие метрики в хранилище
func (storage *MemStorage) GetGauge(name string) (Gauge, bool) {
	value, ok := storage.gauge[name]
	return value, ok
}

func (storage *MemStorage) GetCounter(name string) (Counter, bool) {
	cValue, ok := storage.counter[name]
	return cValue, ok
}

func NewMemStorage() Storage {
	return &MemStorage{
		//metrics: make(map[string]interface{}),
		gauge:   make(map[string]Gauge),
		counter: make(map[string]Counter),
	}
}

// MeStore Хранилище метрик в памяти.
var MeStore Storage

// Инициализация хранилища
func init() {
	MeStore = NewMemStorage()
}

func (storage *MemStorage) GetGauges() map[string]Gauge {
	return storage.gauge
}

func (storage *MemStorage) GetCounters() map[string]Counter {
	return storage.counter
}
