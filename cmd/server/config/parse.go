package config

import (
	"encoding/json"
	"errors"
	"flag"
	incnf "gmetrics/internal/config"
	"os"

	"github.com/caarlos0/env/v6"
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

	// Заполняем из файла то, что не заполнили из остальных
	if err := parseFromFile(cnf); err != nil {
		return nil, err
	}

	key, err := incnf.ParsePrivateKeyFromFile(cnf.CryptoKeyPath)
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
		params.StoreInterval = cnf.StoreInterval
	}
	if cnf.DatabaseDSN != "" {
		params.DatabaseDSN = cnf.DatabaseDSN
	}
	if cnf.HashKey != "" {
		params.HashKey = cnf.HashKey
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
func parseFromCli(cnf *CliConfig) (parseError error) {
	// Регистрируем флаги конфигурации
	flag.StringVar(&cnf.Address, "a", DefaultServerURL, "address and port to run server")
	flag.StringVar(&cnf.LogLevel, "ll", DefaultLogLevel, "level of logging")
	flag.StringVar(&cnf.FileStorage, "f", DefaultFilePath, "store file path")
	flag.StringVar(&cnf.DatabaseDSN, "d", DefaultDatabaseDSN, "database connection")
	flag.Int64Var(&cnf.StoreInterval, "i", DefaultStoreInterval, "frequency of save metrics. 0 is sync mode")
	flag.BoolVar(&cnf.Restore, "r", DefaultRestore, "need to restore")
	flag.StringVar(&cnf.HashKey, "k", DefaultHashKey, "encrypted key")
	flag.StringVar(&cnf.CryptoKeyPath, "crypto-key", "", "crypto key")
	flag.StringVar(&cnf.ConfigFilePath, "c", "", "Path to the configuration file (shorthand)")
	flag.StringVar(&cnf.ConfigFilePath, "config", "", "Path to the configuration file")

	// Парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse() // Сейчас будет выход из приложения, поэтому код ниже не будет исполнен, но может пригодиться в будущем, если поменять флаг выхода или будет несколько сетов
	if !flag.Parsed() {
		return errors.New("error while parse flags")
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
	if fileConf.Address != "" && cnf.Address == DefaultServerURL {
		cnf.Address = fileConf.Address
	}
	if fileConf.StoreInterval.Duration != 0 && cnf.StoreInterval == DefaultStoreInterval {
		cnf.StoreInterval = int64(fileConf.StoreInterval.Seconds())
	}
	if fileConf.Restore != DefaultRestore && cnf.Restore == DefaultRestore {
		cnf.Restore = fileConf.Restore
	}
	if fileConf.StoreFile != "" && cnf.FileStorage == DefaultFilePath {
		cnf.FileStorage = fileConf.StoreFile
	}
	if fileConf.DatabaseDsn != "" && cnf.DatabaseDSN == DefaultDatabaseDSN {
		cnf.DatabaseDSN = fileConf.DatabaseDsn
	}
	if fileConf.CryptoKey != "" && cnf.CryptoKeyPath == "" {
		cnf.CryptoKeyPath = fileConf.CryptoKey
	}
	return nil
}
