package main

import (
	"context"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/collector/sender"
	"gmetrics/cmd/agent/config"
	"gmetrics/internal/logger"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	log.Println("Agent is starting")
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
	)

	// Создаём новую коллекцию метрик и устанавливаем её глобально
	collection.Collection = collection.NewCollection()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Создаём группу ожидания на 2 потока: сборки данных и отправки
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		collection.CollectProcess(ctx)
		defer wg.Done()
	}() // Запускаем сборку данных

	// Запускаем отправку данных
	client := sender.New(collection.Collection)
	logger.Log.Info("New sender client created")
	go func() {
		client.PeriodicSender(ctx)
		defer wg.Done()
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
	lgr, err := logger.New(config.Params.LogLevel)
	if err != nil {
		return nil, err
	}
	logger.Log = lgr

	return lgr, nil
}
