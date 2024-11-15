package config

import "crypto/rsa"

const (
	// DefaultServerURL Url сервера получателя метрик по умолчанию
	DefaultServerURL = "localhost:8080"

	// DefaultLogLevel Уровень логирования по умолчанию
	DefaultLogLevel = "info"

	// DefaultFilePath путь хранения метрик по умолчанию
	DefaultFilePath = "storage.json"

	// DefaultStoreInterval период сохранения метрик в файл по умолчанию
	DefaultStoreInterval int64 = 300

	// DefaultRestore надобность загрузки старых данных из файла при включении
	DefaultRestore = false

	// DefaultDatabaseDSN подключение к базе данных
	DefaultDatabaseDSN = "" //"postgresql://postgres:example@127.0.0.1:5432/gmetrics"

	// DefaultHashKey ключ шифрования по умолчанию
	DefaultHashKey = ""
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	// Address адрес сервера
	Address     string `env:"ADDRESS"`
	LogLevel    string `env:"LOG_LEVEL"`    // Уровень логирования
	FileStorage string `env:"FILE_STORAGE"` // Путь к хранению файлов, если не указан, то будет создано обычное хранилище в памяти
	DatabaseDSN string `env:"DATABASE_DSN"` // подключение к базе данных
	// HashKey Ключ для шифрования
	HashKey       string `env:"KEY"`
	StoreInterval int64  `env:"STORE_INTERVAL"` // период сохранения метрик в файл; 0 - синхронный режим
	Restore       bool   `env:"RESTORE"`        // Надобность загрузки старых данных из файла при включении
	// CryptoKeyPath Путь к файлу с приватным ключом
	CryptoKeyPath string `env:"CRYPTO_KEY"`
	// CryptoKey Приватный ключ для дешифрования тела запроса
	CryptoKey *rsa.PrivateKey
	// ConfigFilePath Путь к файлу с конфигурацией
	ConfigFilePath string `env:"CONFIG"`
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
		HashKey:       DefaultHashKey,
	}
}
