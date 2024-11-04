package main

import (
	"context"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/collector/sender"
	"gmetrics/cmd/agent/config"
	"gmetrics/cmd/agent/sendpool"
	"gmetrics/internal/buildflags"
	"gmetrics/internal/logger"
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	buildflags.PrintBuildInformation()
	log.Println("Agent is starting")
	go func() {
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()
	// Устанавливаем настройки
	cnf, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	config.Params = cnf
	_, err = InitLogger()
	if err != nil {
		log.Fatal(err)
	}
	// Показываем конфигурацию агента
	logger.Log.Infow("Running agent with configuration",
		"poll interval", config.Params.PollInterval,
		"logLevel", config.Params.LogLevel,
		"server url", config.Params.ServerURL,
		"report interval", config.Params.ReportInterval,
		"hash key", config.Params.HashKey,
	)

	// Создаём новую коллекцию метрик и устанавливаем её глобально
	collection.Collection = collection.NewCollection()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Создаём группу ожидания на 2 потока: сборки данных и отправки
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		collection.CollectProcess(ctx)
	}() // Запускаем сборку данных
	go func() {
		defer wg.Done()
		collection.CollectUtilProcess(ctx)
	}() // Запускаем сборку данных использования системы

	// Создаём пул отправок на сервер
	sendPool := sendpool.New(ctx, config.Params.RateLimit, config.Params.HashKey, config.Params.ServerURL)

	// Запускаем отправку данных
	client := sender.New(collection.Collection, sendPool)
	logger.Log.Info("New sender client created")
	go func() {
		defer wg.Done()
		client.PeriodicSender(ctx)
	}()

	// Ожидаем сигнала завершения Ctrl+C, чтобы корректно завершить работу
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
		cancel()
	case <-ctx.Done():
		// continue
	}
	logger.Log.Info("Agent is stopping")
	wg.Wait() // Ожидаем завершения всех горутин
	logger.Log.Infow("Agent stopped")
}

// InitLogger инициализируем логер
func InitLogger() (*zap.SugaredLogger, error) {
	loggerLevel, err := logger.ParseLevel(config.Params.LogLevel)
	if err != nil {
		return nil, err
	}
	lgr, err := logger.New(loggerLevel)
	if err != nil {
		return nil, err
	}
	logger.Log = lgr

	return lgr, nil
}
