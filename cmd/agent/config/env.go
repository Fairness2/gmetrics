package config

import (
	"fmt"
)

const (
	// DefaultPollInterval Интервал между сборкой данных по умолчанию
	DefaultPollInterval int64 = 2

	// DefaultReportInterval Интервал между отправкой данных по умолчанию
	DefaultReportInterval int64 = 10

	// DefaultServerURL Url сервера получателя метрик по умолчанию
	DefaultServerURL = "http://localhost:8080"

	// DefaultLogLevel Уровень логирования по умолчанию
	DefaultLogLevel = "info"
)

// CliConfig конфигурация клиента из командной строки
type CliConfig struct {
	// PollInterval Интервал между сборкой данных
	PollInterval int64 `env:"POLL_INTERVAL"`
	// ReportInterval Интервал между отправкой данных
	ReportInterval int64 `env:"REPORT_INTERVAL"`
	// ServerURL Url сервера получателя метрик
	ServerURL string `env:"ADDRESS"`
	// Уровень логирования
	LogLevel string `env:"LOG_LEVEL"`
}

// Params конфигурация приложения
var Params *CliConfig

// InitializeDefaultConfig инициализация конфигурации приложения
func InitializeDefaultConfig() *CliConfig {
	return &CliConfig{
		PollInterval:   DefaultPollInterval,
		ReportInterval: DefaultReportInterval,
		ServerURL:      DefaultServerURL,
		LogLevel:       DefaultLogLevel,
	}
}

// PrintConfig возвращает строку с информацией о текущей конфигурации сервера и интервалах сбора метрик и отправки метрик.
// Server Address: <адрес сервера>
// Frequency of metrics collection: <интервал сбора метрик> s.
// Frequency of sending metrics: <интервал отправки метрик> s.
func PrintConfig(cnf *CliConfig) string {
	return fmt.Sprintf("Server Address: %s\nFrequency of metrics collection: %d s.\nFrequency of sending metrics: %d s.\n", cnf.ServerURL, cnf.PollInterval, cnf.ReportInterval)
}
