package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"strconv"
	"strings"
	"time"
)

// PollInterval Интервал между сборкой данных
var PollInterval = time.Duration(2) * time.Second

// ReportInterval Интервал между отправкой данных
var ReportInterval = time.Duration(10) * time.Second

// ServerURL Url сервера получателя метрик
var ServerURL = "http://localhost:8080"

// Parse инициализирует новую консольную конфигурацию, обрабатывает аргументы командной строки
func Parse() {
	parseFromCli()
	parseFromEnv()
}

// PrintConfig возвращает строку с информацией о текущей конфигурации сервера и интервалах сбора метрик и отправки метрик.
// Server Address: <адрес сервера>
// Frequency of metrics collection: <интервал сбора метрик> s.
// Frequency of sending metrics: <интервал отправки метрик> s.
func PrintConfig() string {
	return fmt.Sprintf("Server Address: %s\nFrequency of metrics collection: %d s.\nFrequency of sending metrics: %d s.\n", ServerURL, PollInterval/time.Second, ReportInterval/time.Second)
}

// parseFromEnv заполняем конфигурацию переменных из окружения
func parseFromEnv() {
	cnf := struct {
		PollInterval   int    `env:"POLL_INTERVAL" envDefault:"-2"`
		ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"-2"`
		ServerURL      string `env:"ADDRESS"`
	}{}
	err := env.Parse(&cnf)
	// Если ошибка, то считаем, что вывести конфигурацию из окружения не удалось
	if err != nil {
		log.Print(err)
		return
	}
	if cnf.PollInterval > 0 {
		PollInterval = time.Duration(cnf.PollInterval) * time.Second
	}
	if cnf.ReportInterval > 0 {
		ReportInterval = time.Duration(cnf.ReportInterval) * time.Second
	}
	if cnf.ServerURL != "" {
		_ = setServerURL(cnf.ServerURL)
	}
}

// parseFromCli заполняем конфигурацию из параметров командной строки
func parseFromCli() {
	// Регистрируем флаги конфигурации
	flag.Func("a", "server address and port", setServerURL)
	flag.Func("p", "frequency of metrics collection", func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return errors.New("invalid frequency of metrics collection: " + err.Error())
		}
		PollInterval = time.Duration(val) * time.Second
		return nil
	})
	flag.Func("r", "frequency of sending metrics", func(s string) error {
		val, err := strconv.Atoi(s)
		if err != nil {
			return errors.New("invalid frequency of sending metrics: " + err.Error())
		}
		ReportInterval = time.Duration(val) * time.Second
		return nil
	})

	// Парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}

func setServerURL(s string) error {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		ServerURL = s
	} else {
		ServerURL = "http://" + s
	}

	return nil
}
