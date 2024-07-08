package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Duration time.Duration

func (d Duration) GetInSeconds() time.Duration {
	return time.Duration(d) * time.Second
}

// PollInterval Интервал между сборкой данных
var PollInterval = time.Duration(2) * time.Second

// ReportInterval Интервал между отправкой данных
var ReportInterval = time.Duration(10) * time.Second

// ServerURL Url сервера получателя метрик
var ServerURL = "http://localhost:8080"

// Parse инициализирует новую консольную конфигурацию, обрабатывает аргументы командной строки
func Parse() {
	// Регистрируем флаги конфигурации
	flag.Func("a", "server address and port", func(s string) error {
		if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
			ServerURL = s
		} else {
			ServerURL = "http://" + s
		}

		return nil
	})
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

func PrintConfig() string {
	return fmt.Sprintf("Server Address: %s\nFrequency of metrics collection: %d s.\nFrequency of sending metrics: %d s.\n", ServerURL, PollInterval/time.Second, ReportInterval/time.Second)
}
