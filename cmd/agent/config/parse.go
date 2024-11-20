package config

import (
	"encoding/json"
	"errors"
	"flag"
	incnf "gmetrics/internal/config"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
)

// Parse инициализирует новую консольную конфигурацию, обрабатывает аргументы командной строки
func Parse() (*CliConfig, error) {
	// Регистрируем новое хранилище
	cnf := InitializeDefaultConfig()
	// Заполняем конфигурацию из параметров командной строки
	if err := parseFromCli(cnf); err != nil {
		return nil, err
	}
	// Заполняем конфигурацию из окружения
	if err := parseFromEnv(cnf); err != nil {
		return nil, err
	}
	// Заполняем из файла то, что не заполнили из остальных
	if err := parseFromFile(cnf); err != nil {
		return nil, err
	}

	key, err := incnf.ParsePublicKeyFromFile(cnf.CryptoKeyPath)
	if err != nil && !errors.Is(err, incnf.ErrorEmptyKeyPath) {
		return nil, err
	}
	if key != nil {
		cnf.CryptoKey = key
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
	if cnf.LogLevel != "" {
		params.LogLevel = cnf.LogLevel
	}
	if cnf.HashKey != "" {
		params.HashKey = cnf.HashKey
	}
	if cnf.RateLimit > 0 {
		params.RateLimit = cnf.RateLimit
	}
	if cnf.CryptoKeyPath != "" {
		params.CryptoKeyPath = cnf.CryptoKeyPath
	}
	if cnf.ConfigFilePath != "" {
		params.ConfigFilePath = cnf.ConfigFilePath
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
	flag.StringVar(&cnf.LogLevel, "ll", DefaultLogLevel, "level of logging")
	flag.StringVar(&cnf.HashKey, "k", DefaultHashKey, "encrypted key")
	flag.IntVar(&cnf.RateLimit, "l", DefaultRateLimit, "number of simultaneously outgoing requests to the server")
	flag.StringVar(&cnf.CryptoKeyPath, "crypto-key", "", "crypto key")
	flag.StringVar(&cnf.ConfigFilePath, "c", "", "Path to the configuration file (shorthand)")
	flag.StringVar(&cnf.ConfigFilePath, "config", "", "Path to the configuration file")

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

// parseFromFile заполняем конфигурацию из файла конфигурации
func parseFromFile(cnf *CliConfig) error {
	if cnf.ConfigFilePath == "" {
		return nil
	}
	file, err := os.ReadFile(cnf.ConfigFilePath)
	if err != nil {
		return err
	}
	var fileConf FileConfig
	if err := json.Unmarshal(file, &fileConf); err != nil {
		return err
	}
	if fileConf.Address != "" && cnf.ServerURL == DefaultServerURL {
		if err = setServerURL(fileConf.Address, cnf); err != nil {
			return err
		}
	}
	if fileConf.ReportInterval.Duration != 0 && cnf.ReportInterval == DefaultReportInterval {
		cnf.ReportInterval = int64(fileConf.ReportInterval.Seconds())
	}
	if fileConf.PollInterval.Duration != 0 && cnf.PollInterval == DefaultPollInterval {
		cnf.PollInterval = int64(fileConf.PollInterval.Seconds())
	}
	if fileConf.CryptoKey != "" && cnf.CryptoKeyPath == "" {
		cnf.CryptoKeyPath = fileConf.CryptoKey
	}
	return nil
}
