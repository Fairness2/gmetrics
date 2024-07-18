package config

import (
	"fmt"
	"time"
)

// CliConfig конфигурация клиента из командной строки
type CliConfig struct {
	// PollInterval Интервал между сборкой данных
	PollInterval time.Duration
	// ReportInterval Интервал между отправкой данных
	ReportInterval time.Duration
	// ServerURL Url сервера получателя метрик
	ServerURL string
}

// Params конфигурация приложения
var Params *CliConfig

// InitializeNewCliConfig инициализация конфигурации приложения
func InitializeNewCliConfig() *CliConfig {
	return &CliConfig{
		PollInterval:   DefaultPollInterval,
		ReportInterval: DefaultReportInterval,
		ServerURL:      DefaultServerURL,
	}
}

// DefaultPollInterval Интервал между сборкой данных по умолчанию
var DefaultPollInterval = 2 * time.Second

// DefaultReportInterval Интервал между отправкой данных по умолчанию
var DefaultReportInterval = 10 * time.Second

// DefaultServerURL Url сервера получателя метрик по умолчанию
var DefaultServerURL = "http://localhost:8080"

// PrintConfig возвращает строку с информацией о текущей конфигурации сервера и интервалах сбора метрик и отправки метрик.
// Server Address: <адрес сервера>
// Frequency of metrics collection: <интервал сбора метрик> s.
// Frequency of sending metrics: <интервал отправки метрик> s.
func PrintConfig(cnf *CliConfig) string {
	return fmt.Sprintf("Server Address: %s\nFrequency of metrics collection: %d s.\nFrequency of sending metrics: %d s.\n", cnf.ServerURL, cnf.PollInterval/time.Second, cnf.ReportInterval/time.Second)
}
