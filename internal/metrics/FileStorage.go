package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"gmetrics/internal/contextkeys"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics/fileworker"
	"io"
	"os"
	"time"
)

// Writer интерфейс записывающего в файл типа
type Writer interface {
	Write(v any) error
	io.Closer
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

// Sync синхронизация данных хранилища в файл по таймеру
func (storage *DurationFileStorage) Sync(ctx context.Context) {
	interval := ctx.Value(contextkeys.SyncInterval).(time.Duration)
	logger.Log.Infof("Sync metrics process starts. Period is %d seconds", interval/time.Second)
	ticker := time.NewTicker(interval)
	for {
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-ticker.C:
			logger.Log.Debug("Sync metrics")
			if err := storage.Flush(); err != nil {
				logger.Log.Error(err)
			}
		case <-ctx.Done():
			logger.Log.Debug("Sync metrics before end")
			ticker.Stop()
			if err := storage.Flush(); err != nil {
				logger.Log.Error(err)
			}
			logger.Log.Debug("Synced")
			return
		}
	}
}

// Close Закрытие писателя (файла)
func (storage *DurationFileStorage) Close() error {
	return storage.writer.Close()
}

// FlushAndClose синхронизация данных и закрытие писателя (файла)
func (storage *DurationFileStorage) FlushAndClose() error {
	if err := storage.Flush(); err != nil {
		return err
	}
	return storage.writer.Close()
}

// SetGauge переопределённый метод с записью в файл в случае синхронного режима
func (storage *DurationFileStorage) SetGauge(name string, value Gauge) {
	storage.Storage.SetGauge(name, value)
	if storage.SyncMode {
		if err := storage.Flush(); err != nil {
			logger.Log.Error(err)
		}
	}

}

// AddCounter переопределённый метод с записью в файл в случае синхронного режима
func (storage *DurationFileStorage) AddCounter(name string, value Counter) {
	storage.Storage.AddCounter(name, value)
	if storage.SyncMode {
		if err := storage.Flush(); err != nil {
			logger.Log.Error(err)
		}
	}
}

// NewFileStorage создание нового хранилища
// filename - имя файла
// restore - нужно ли загрузить инициализирующие данные из файла
func NewFileStorage(filename string, restore bool, syncMode bool) (*DurationFileStorage, error) {
	storage := NewMemStorage()
	if restore {
		// Восстанавливаем хранилище из файла, возвращаем ошибку если чтение вернуло ошибку не с типом несуществующего файла или пустого файла
		if err := restoreFromFile(filename, storage); err != nil {
			logger.Log.Infow("Restore store failed", "filename", filename, "error", err)
			if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, io.EOF) {
				return nil, err
			}
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
