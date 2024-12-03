package config

import (
	"crypto/rsa"
	"net"
)

const (
	// DefaultServerURL Url сервера получателя метрик по умолчанию
	DefaultServerURL = "localhost:8080"
	// DefaultRPCServerURL Url сервера получателя метрик по rpc по умолчанию
	DefaultRPCServerURL = "localhost:8675"

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
	Address          string          `env:"ADDRESS"`        // адрес сервера
	RPCAddress       string          `env:"RPC_ADDRESS"`    // адрес сервера
	LogLevel         string          `env:"LOG_LEVEL"`      // Уровень логирования
	FileStorage      string          `env:"FILE_STORAGE"`   // Путь к хранению файлов, если не указан, то будет создано обычное хранилище в памяти
	DatabaseDSN      string          `env:"DATABASE_DSN"`   // подключение к базе данных
	HashKey          string          `env:"KEY"`            // Ключ для шифрования
	StoreInterval    int64           `env:"STORE_INTERVAL"` // период сохранения метрик в файл; 0 - синхронный режим
	Restore          bool            `env:"RESTORE"`        // Надобность загрузки старых данных из файла при включении
	CryptoKeyPath    string          `env:"CRYPTO_KEY"`     // Путь к файлу с приватным ключом
	CryptoKey        *rsa.PrivateKey // Приватный ключ для дешифрования тела запроса
	ConfigFilePath   string          `env:"CONFIG"`         // Путь к файлу с конфигурацией
	TrustedSubnetStr string          `env:"TRUSTED_SUBNET"` // CIDR адрес подсети, запросы из которого будут обрабатываться
	TrustedSubnet    *net.IPNet      // Доверенная подсеть
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
		RPCAddress:    DefaultRPCServerURL,
	}
}
