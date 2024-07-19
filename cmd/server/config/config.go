package config

import (
	"fmt"
	"go.uber.org/zap"
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	// Address адрес сервера
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

// DefaultServerURL Url сервера получателя метрик по умолчанию
var DefaultServerURL = "localhost:8080"

// DefaultLogLevel Уровень логирования по умолчанию
var DefaultLogLevel = zap.InfoLevel.String()

// Params конфигурация приложения
var Params *CliConfig

// InitializeDefaultConfig инициализация конфигурации приложения
func InitializeDefaultConfig() *CliConfig {
	return &CliConfig{
		Address:  DefaultServerURL,
		LogLevel: DefaultLogLevel,
	}
}

// PrintConfig возвращает строку с информацией о текущей конфигурации сервера и интервалах сбора метрик и отправки метрик.
func PrintConfig(cnf *CliConfig) string {
	return fmt.Sprintf("Server Address: %s, Log level: %s\n", cnf.Address, cnf.LogLevel)
}
