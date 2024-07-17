package config

import (
	"fmt"
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	// Address адрес сервера
	Address string `env:"ADDRESS"`
}

// DefaultServerURL Url сервера получателя метрик по умолчанию
var DefaultServerURL = "localhost:8080"

// Params конфигурация приложения
var Params *CliConfig

// InitializeNewCliConfig инициализация конфигурации приложения
func InitializeNewCliConfig() *CliConfig {
	return &CliConfig{
		Address: DefaultServerURL,
	}
}

// SetGlobalConfig устанавливает глобальную конфигурацию приложения
func SetGlobalConfig(cnf *CliConfig) {
	Params = cnf
}

// PrintConfig возвращает строку с информацией о текущей конфигурации сервера и интервалах сбора метрик и отправки метрик.
func PrintConfig(cnf *CliConfig) string {
	return fmt.Sprintf("Server Address: %s\n", cnf.Address)
}
