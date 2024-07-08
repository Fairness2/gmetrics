package config

import (
	"flag"
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	// Address адрес сервера
	Address string
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

	// Регистрируем флаги конфигурации
	flag.StringVar(&Params.Address, "a", "localhost:8080", "address and port to run server")

	// Парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}
