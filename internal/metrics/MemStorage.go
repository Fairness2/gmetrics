package metrics

import "sync"

// MemStorage Хранилище метрик в памяти
type MemStorage struct {
	gauge   map[string]Gauge
	counter map[string]Counter
	mutex   *sync.RWMutex
	//metrics map[string]any
}

// Set Установка значения метрики в хранилище
// Parameters:
//
//	name: the name of the metric
//	value: the value to be stored for the metric
/*func (storage *MemStorage) Set(name string, value any) {
	storage.metrics[name] = value
}*/

func (storage *MemStorage) SetGauge(name string, value Gauge) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	storage.gauge[name] = value
}

func (storage *MemStorage) AddCounter(name string, value Counter) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
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
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	value, ok := storage.gauge[name]
	return value, ok
}

func (storage *MemStorage) GetCounter(name string) (Counter, bool) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	cValue, ok := storage.counter[name]
	return cValue, ok
}

func NewMemStorage() Storage {
	return &MemStorage{
		//metrics: make(map[string]any),
		gauge:   make(map[string]Gauge),
		counter: make(map[string]Counter),
		mutex:   new(sync.RWMutex),
	}
}

func (storage *MemStorage) GetGauges() map[string]Gauge {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	return storage.gauge
}

func (storage *MemStorage) GetCounters() map[string]Counter {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	return storage.counter
}
