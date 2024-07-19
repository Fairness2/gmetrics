package main

import (
	"context"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/collector/sender"
	"gmetrics/cmd/agent/config"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	log.Println("Agent is starting")
	// Устанавливаем настройки
	cnf, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	config.Params = cnf
	log.Print(config.PrintConfig(cnf))

	// Создаём новую коллекцию метрик и устанавливаем её глобально
	collection.Collection = collection.NewCollection()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Таймер, чтобы ограничить время работы агента, чтобы пройти тесты
	defer cancel()
	// Создаём группу ожидания на 2 потока: сборки данных и отправки
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		collection.CollectProcess(ctx)
		defer wg.Done()
	}() // Запускаем сборку данных

	// Запускаем отправку данных
	client := sender.NewSender(collection.Collection)
	log.Println("New sender client created")
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
	log.Println("Agent is stopping")
	wg.Wait() // Ожидаем завершения всех горутин
	log.Println("Agent stopped")
}
