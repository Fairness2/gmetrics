package config

import (
	"flag"
	"os"
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	// Address адрес сервера
	Address string `env:"ADDRESS"`
}

// Config конфигурация приложения
var Params *CliConfig

// InitializeNewCliConfig инициализация конфигурации приложения
func InitializeNewCliConfig() {
	Params = new(CliConfig)
}

// Parse инициализирует новую консольную конфигурацию, обрабатывает аргументы командной строки
func Parse() {
	// Регистрируем новое хранилище
	InitializeNewCliConfig()
	// Заполняем конфигурацию из параметров командной строки
	parseFromCli()
	// Заполняем конфигурацию из окружения
	parseFromEnv()
}

// parseFromEnv заполняем конфигурацию переменных из окружения
func parseFromEnv() {
	envAddr := os.Getenv("ADDRESS")
	if envAddr != "" {
		Params.Address = envAddr
	}
}

// parseFromCli заполняем конфигурацию из параметров командной строки
func parseFromCli() {
	// Регистрируем флаги конфигурации
	flag.StringVar(&Params.Address, "a", "localhost:8080", "address and port to run server")
	// Парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}
