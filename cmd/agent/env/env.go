package env

import "time"

// PollInterval Интервал между сборкой данных
var PollInterval = time.Duration(2) * time.Second

// ReportInterval Интервал между отправкой данных
var ReportInterval = time.Duration(10) * time.Second

// ServerURL Url сервера получателя метрик
var ServerURL = "http://localhost:8080"
