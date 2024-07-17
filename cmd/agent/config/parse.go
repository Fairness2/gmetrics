package config

import (
	"errors"
	"flag"
	"github.com/caarlos0/env/v6"
	"strconv"
	"strings"
	"time"
)

// Parse инициализирует новую консольную конфигурацию, обрабатывает аргументы командной строки
func Parse() (*CliConfig, error) {
	// Регистрируем новое хранилище
	cnf := InitializeNewCliConfig()
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
		PollInterval   time.Duration `env:"POLL_INTERVAL"`
		ReportInterval time.Duration `env:"REPORT_INTERVAL"`
		ServerURL      string        `env:"ADDRESS"`
	}{}
	err := env.Parse(&cnf)
	// Если ошибка, то считаем, что вывести конфигурацию из окружения не удалось
	if err != nil {
		return err
	}
	if cnf.PollInterval > 0 {
		params.PollInterval = cnf.PollInterval * time.Second
	}
	if cnf.ReportInterval > 0 {
		params.ReportInterval = cnf.ReportInterval * time.Second
	}
	if cnf.ServerURL != "" {
		if err = setServerURL(cnf.ServerURL, params); err != nil {
			return err
		}
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
	flag.Func("p", "frequency of metrics collection", func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			parseError = errors.New("invalid frequency of metrics collection: " + err.Error())
			return parseError
		}
		cnf.PollInterval = time.Duration(val) * time.Second
		return nil
	})
	flag.Func("r", "frequency of sending metrics", func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			parseError = errors.New("invalid frequency of sending metrics: " + err.Error())
			return parseError
		}
		cnf.ReportInterval = time.Duration(val) * time.Second
		return nil
	})

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
