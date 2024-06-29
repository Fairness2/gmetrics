package metrics

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
//
// TODO Знаю, что сейчас при наличии одинаковых имён будет передан гауге, поправлю в дальнейшем, когда гет понадобится
func (storage *MemStorage) Get(name string) (interface{}, bool) {
	//value, ok := storage.metrics[name]
	value, ok := storage.gauge[name]
	if ok {
		return value, ok

	}
	cValue, ok := storage.counter[name]
	return cValue, ok
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		//metrics: make(map[string]interface{}),
		gauge:   make(map[string]Gauge),
		counter: make(map[string]Counter),
	}
}

// MeStore Хранилище метрик в памяти.
var MeStore *MemStorage

// Инициализация хранилища
func init() {
	MeStore = NewMemStorage()
}
