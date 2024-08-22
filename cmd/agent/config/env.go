package config

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
