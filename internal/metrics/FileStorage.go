package metrics

import (
	"context"
	"encoding/json"
	"gmetrics/internal/contextkeys"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics/fileworker"
	"os"
	"time"
)

// Writer интерфейс записывающего в файл типа
type Writer interface {
	Write(v any) error
	Close() error
}

// DurationFileStorage хранилище с циклической записью в файл данных
type DurationFileStorage struct {
	Storage
	writer   Writer
	SyncMode bool // Флаг синхронного режима, в нём после записи в хранилище, сразу оно будет записано в файл
}

// Flush запись данных в файл
func (storage *DurationFileStorage) Flush() error {
	return storage.writer.Write(storage.Storage)
}

func (storage *DurationFileStorage) Sync(ctx context.Context) {
	interval := ctx.Value(contextkeys.SyncInterval).(time.Duration)
	logger.G.Infof("Sync metrics process starts. Period is %d seconds", interval/time.Second)
	for {
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-time.After(interval):
			logger.G.Debug("Sync metrics")
			if err := storage.Flush(); err != nil {
				logger.G.Error(err)
			}
		case <-ctx.Done():
			logger.G.Debug("Sync metrics before end")
			if err := storage.Flush(); err != nil {
				logger.G.Error(err)
			}
			return
		}
	}
}

// Close Закрытие писателя (файла)
func (storage *DurationFileStorage) Close() error {
	return storage.writer.Close()
}

// SetGauge переопределённый метод с записью в файл в случае синхронного режима
func (storage *DurationFileStorage) SetGauge(name string, value Gauge) {
	storage.Storage.SetGauge(name, value)
	if storage.SyncMode {
		if err := storage.Flush(); err != nil {
			logger.G.Error(err)
		}
	}

}

// AddCounter переопределённый метод с записью в файл в случае синхронного режима
func (storage *DurationFileStorage) AddCounter(name string, value Counter) {
	storage.Storage.AddCounter(name, value)
	if storage.SyncMode {
		if err := storage.Flush(); err != nil {
			logger.G.Error(err)
		}
	}
}

// NewFileStorage создание нового хранилища
// filename - имя файла
// restore - нужно ли загрузить инициализирующие данные из файла
func NewFileStorage(filename string, restore bool, syncMode bool) (*DurationFileStorage, error) {
	storage := NewMemStorage()
	if restore {
		if err := restoreFromFile(filename, storage); err != nil {
			return nil, err
		}
	}
	writer, err := fileworker.NewWriter(filename)
	if err != nil {
		return nil, err
	}
	return &DurationFileStorage{
		Storage:  storage,
		writer:   writer,
		SyncMode: syncMode,
	}, nil
}

// restoreFromFile чтение данных из файла при инициализации хранилища
func restoreFromFile(filename string, storage Storage) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	return decoder.Decode(storage)
}
