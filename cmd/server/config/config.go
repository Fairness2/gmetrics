package config

import (
	"fmt"
	"go.uber.org/zap"
	"time"
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	// Address адрес сервера
	Address       string        `env:"ADDRESS"`
	LogLevel      string        `env:"LOG_LEVEL"`      // Уровень логирования
	FileStorage   string        `env:"FILE_STORAGE"`   // Путь к хранению файлов, если не указан, то будет создано обычное хранилище в памяти
	Restore       bool          `env:"RESTORE"`        // Надобность загрузки старых данных из файла при включении
	StoreInterval time.Duration `env:"STORE_INTERVAL"` // период сохранения метрик в файл; 0 - синхронный режим
}

// DefaultServerURL Url сервера получателя метрик по умолчанию
var DefaultServerURL = "localhost:8080"

// DefaultLogLevel Уровень логирования по умолчанию
var DefaultLogLevel = zap.InfoLevel.String()

// DefaultFilePath путь хранения метрик по умолчанию
var DefaultFilePath = "" //"storage.json"

// DefaultStoreInterval период сохранения метрик в файл по умолчанию
var DefaultStoreInterval = 300 * time.Second

// DefaultRestore надобность загрузки старых данных из файла при включении
var DefaultRestore = false

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
	}
}

// PrintConfig возвращает строку с информацией о текущей конфигурации сервера и интервалах сбора метрик и отправки метрик.
func PrintConfig(cnf *CliConfig) string {
	return fmt.Sprintf("Server Address: %s, Log level: %s\n", cnf.Address, cnf.LogLevel)
}
