package metrics

import "sync"

// MemStorage Хранилище метрик в памяти
type MemStorage struct {
	Gauge   map[string]Gauge   `json:"gauge"`
	Counter map[string]Counter `json:"counter"`
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

func (storage *MemStorage) SetGauge(name string, value Gauge) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	storage.Gauge[name] = value
	return nil
}

func (storage *MemStorage) AddCounter(name string, value Counter) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	oldValue, ok := storage.Counter[name]
	if ok {
		value = oldValue.Add(value)
	}
	storage.Counter[name] = value
	return nil
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
	value, ok := storage.Gauge[name]
	return value, ok
}

func (storage *MemStorage) GetCounter(name string) (Counter, bool) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	cValue, ok := storage.Counter[name]
	return cValue, ok
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		//metrics: make(map[string]any),
		Gauge:   make(map[string]Gauge),
		Counter: make(map[string]Counter),
		mutex:   new(sync.RWMutex),
	}
}

func (storage *MemStorage) GetGauges() (map[string]Gauge, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	return storage.Gauge, nil
}

func (storage *MemStorage) GetCounters() (map[string]Counter, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	return storage.Counter, nil
}
