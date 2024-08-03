package config

import (
	"fmt"
	"time"
)

const (
	// DefaultServerURL Url сервера получателя метрик по умолчанию
	DefaultServerURL = "localhost:8080"

	// DefaultLogLevel Уровень логирования по умолчанию
	DefaultLogLevel = "info"

	// DefaultFilePath путь хранения метрик по умолчанию
	DefaultFilePath = "storage.json"

	// DefaultStoreInterval период сохранения метрик в файл по умолчанию
	DefaultStoreInterval = 300 * time.Second

	// DefaultRestore надобность загрузки старых данных из файла при включении
	DefaultRestore = false

	// DefaultDatabaseDSN подключение к базе данных
	DefaultDatabaseDSN = "postgresql://postgres:example@127.0.0.1:5432/gmetrics"
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	// Address адрес сервера
	Address       string        `env:"ADDRESS"`
	LogLevel      string        `env:"LOG_LEVEL"`      // Уровень логирования
	FileStorage   string        `env:"FILE_STORAGE"`   // Путь к хранению файлов, если не указан, то будет создано обычное хранилище в памяти
	Restore       bool          `env:"RESTORE"`        // Надобность загрузки старых данных из файла при включении
	StoreInterval time.Duration `env:"STORE_INTERVAL"` // период сохранения метрик в файл; 0 - синхронный режим
	DatabaseDSN   string        `env:"DATABASE_DSN"`   // подключение к базе данных
}

// Params конфигурация приложения
var Params *CliConfig

// InitializeDefaultConfig инициализация конфигурации приложения
func InitializeDefaultConfig() *CliConfig {
	return &CliConfig{
		Address:       DefaultServerURL,
		LogLevel:      DefaultLogLevel,
		FileStorage:   DefaultFilePath,
		Restore:       DefaultRestore,
		StoreInterval: DefaultStoreInterval,
		DatabaseDSN:   DefaultDatabaseDSN,
	}
}

// PrintConfig возвращает строку с информацией о текущей конфигурации сервера и интервалах сбора метрик и отправки метрик.
func PrintConfig(cnf *CliConfig) string {
	return fmt.Sprintf("Server Address: %s, Log level: %s\n", cnf.Address, cnf.LogLevel)
}
