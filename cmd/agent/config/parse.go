package config

import (
	"errors"
	"flag"
	"github.com/caarlos0/env/v6"
	"strings"
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
	cnf := CliConfig{}
	err := env.Parse(&cnf)
	// Если ошибка, то считаем, что вывести конфигурацию из окружения не удалось
	if err != nil {
		return err
	}
	if cnf.PollInterval > 0 {
		params.PollInterval = cnf.PollInterval
	}
	if cnf.ReportInterval > 0 {
		params.ReportInterval = cnf.ReportInterval
	}
	if cnf.ServerURL != "" {
		if err = setServerURL(cnf.ServerURL, params); err != nil {
			return err
		}
	}
	if cnf.HashKey != "" {
		params.HashKey = cnf.HashKey
	}

	return nil
}

// parseFromCli заполняем конфигурацию из параметров командной строки
func parseFromCli(cnf *CliConfig) error {
	var parseError error
	// Регистрируем флаги конфигурации
	flag.Func("a", "server address and port", func(s string) error {
		parseError = setServerURL(s, cnf)
		return parseError
	})
	flag.Int64Var(&cnf.PollInterval, "p", DefaultPollInterval, "frequency of metrics collection")
	flag.Int64Var(&cnf.ReportInterval, "r", DefaultReportInterval, "frequency of sending metrics")
	flag.StringVar(&cnf.HashKey, "k", DefaultHashKey, "encrypted key")

	// Парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse() // Сейчас будет выход из приложения, поэтому код ниже не будет исполнен, но может пригодиться в будущем, если поменять флаг выхода или будет несколько сетов
	if !flag.Parsed() {
		if parseError != nil {
			return parseError
		} else {
			return errors.New("error while parse flags")
		}
	}
	return nil
}

// setServerURL задает URL-адрес сервера в параметрах конфигурации.
// Если урл не начинается с "http://" или "https://", то будет дополнен "http://".
// Если установить не удастся или переданный урл некорректен, то будет возвращена ошибка
func setServerURL(s string, cnf *CliConfig) error {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		cnf.ServerURL = s
	} else {
		cnf.ServerURL = "http://" + s
	}

	return nil
}
