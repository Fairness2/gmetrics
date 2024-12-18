package metrics

import "sync"

// MemStorage Хранилище метрик в памяти
type MemStorage struct {
	Gauge   map[string]Gauge   `json:"gauge"`
	Counter map[string]Counter `json:"counter"`
	mutex   *sync.RWMutex
}

// SetGauge устанавливаем gauge
func (storage *MemStorage) SetGauge(name string, value Gauge) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	return storage.unsafeSetGauge(name, value)
}

// unsafeSetGauge устанавливает значение Gauge для данного имени без какой-либо блокировки.
// Предполагается, что вызывающая функция обрабатывает все необходимое управление параллелизмом.
func (storage *MemStorage) unsafeSetGauge(name string, value Gauge) error {
	storage.Gauge[name] = value
	return nil
}

// AddCounter добавляем каунтер
func (storage *MemStorage) AddCounter(name string, value Counter) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	return storage.unsafeAddCounter(name, value)
}

// unsafeAddCounter устанавливает значение Counter для данного имени без какой-либо блокировки.
// Предполагается, что вызывающая функция обрабатывает все необходимое управление параллелизмом.
func (storage *MemStorage) unsafeAddCounter(name string, value Counter) error {
	oldValue, ok := storage.Counter[name]
	if ok {
		value = oldValue.Add(value)
	}
	storage.Counter[name] = value
	return nil
}

// GetGauge получение отдельного gauge
func (storage *MemStorage) GetGauge(name string) (Gauge, bool) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	value, ok := storage.Gauge[name]
	return value, ok
}

// GetCounter получение отдельного counter
func (storage *MemStorage) GetCounter(name string) (Counter, bool) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	cValue, ok := storage.Counter[name]
	return cValue, ok
}

// NewMemStorage создание нового хранилища в памяти
func NewMemStorage() *MemStorage {
	return &MemStorage{
		//metrics: make(map[string]any),
		Gauge:   make(map[string]Gauge),
		Counter: make(map[string]Counter),
		mutex:   new(sync.RWMutex),
	}
}

// GetGauges получение всех gauge
func (storage *MemStorage) GetGauges() (map[string]Gauge, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	newMap := make(map[string]Gauge, len(storage.Gauge))
	for k, v := range storage.Gauge {
		newMap[k] = v
	}
	return newMap, nil
}

// GetCounters получение всех counter
func (storage *MemStorage) GetCounters() (map[string]Counter, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()
	newMap := make(map[string]Counter, len(storage.Counter))
	for k, v := range storage.Counter {
		newMap[k] = v
	}
	return newMap, nil
}

// SetGauges массовое обновление гауге в памяти
func (storage *MemStorage) SetGauges(gauges map[string]Gauge) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	for name, gauge := range gauges {
		if err := storage.unsafeSetGauge(name, gauge); err != nil {
			return err
		}
	}
	return nil
}

// AddCounters массовое обновление каунтер в памяти
func (storage *MemStorage) AddCounters(counters map[string]Counter) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	for name, counter := range counters {
		if err := storage.unsafeAddCounter(name, counter); err != nil {
			return err
		}
	}
	return nil
}
