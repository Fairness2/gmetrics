package config

import (
	"errors"
	"flag"
	"github.com/caarlos0/env/v6"
	"os"
	"strconv"
	"time"
)

// Parse инициализирует новую консольную конфигурацию, обрабатывает аргументы командной строки
func Parse() (*CliConfig, error) {
	// Регистрируем новое хранилище
	cnf := InitializeDefaultConfig()
	// Заполняем конфигурацию из параметров командной строки
	err := parseFromCli(cnf)
	if err != nil {
		return nil, err
	}
	// Заполняем конфигурацию из окружения
	err = parseFromEnv(cnf)
	if err != nil {
		return nil, err
	}
	return cnf, nil
}

// parseFromEnv заполняем конфигурацию переменных из окружения
func parseFromEnv(params *CliConfig) error {
	cnf := struct {
		Address       string `env:"ADDRESS"`
		LogLevel      string `env:"LOG_LEVEL"`
		FileStorage   string `env:"FILE_STORAGE"`
		Restore       bool   `env:"RESTORE"`
		StoreInterval int    `env:"STORE_INTERVAL" envDefault:"-1"`
	}{}
	err := env.Parse(&cnf)
	// Если ошибка, то считаем, что вывести конфигурацию из окружения не удалось
	if err != nil {
		return err
	}
	if cnf.Address != "" {
		params.Address = cnf.Address
	}
	if cnf.LogLevel != "" {
		params.LogLevel = cnf.LogLevel
	}
	if cnf.FileStorage != "" {
		params.FileStorage = cnf.FileStorage
	}
	if _, ok := os.LookupEnv("RESTORE"); ok {
		params.Restore = cnf.Restore
	}
	if cnf.StoreInterval >= 0 {
		params.StoreInterval = time.Duration(cnf.StoreInterval) * time.Second
	}
	return nil
}

// parseFromCli заполняем конфигурацию из параметров командной строки
func parseFromCli(cnf *CliConfig) (parseError error) {
	// Регистрируем флаги конфигурации
	flag.StringVar(&cnf.Address, "a", DefaultServerURL, "address and port to run server")
	flag.StringVar(&cnf.LogLevel, "ll", DefaultLogLevel, "level of logging")
	flag.StringVar(&cnf.FileStorage, "f", DefaultFilePath, "store file path")
	flag.BoolVar(&cnf.Restore, "r", DefaultRestore, "need to restore")
	flag.Func("i", "frequency of save metrics. 0 is sync mode", func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			parseError = errors.New("requency of save metrics: " + err.Error())
			return parseError
		}
		cnf.StoreInterval = time.Duration(val) * time.Second
		return nil
	})
	// Парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse() // Сейчас будет выход из приложения, поэтому код ниже не будет исполнен, но может пригодиться в будущем, если поменять флаг выхода или будет несколько сетов
	if !flag.Parsed() {
		return errors.New("error while parse flags")
	}
	return nil
}
