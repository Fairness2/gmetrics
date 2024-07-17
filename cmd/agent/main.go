package main

import (
	"context"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/collector/sender"
	"gmetrics/cmd/agent/config"
	"log"
	"sync"
	"time"
)

func main() {
	log.Println("Agent is starting")
	// Устанавливаем настройки
	cnf, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	config.SetGlobalConfig(cnf)
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

	// Бесконечный цикл со считыванием ввода консоли, чтобы программа работала пока нужно
	/*for {
		fmt.Println("Agent is running. Print C to finish agent")
		var command string
		_, err := fmt.Fscan(os.Stdin, &command)
		if err != nil {
			log.Println(err)
		}
		if command == "C" {
			cancel()
			break
		}
	}*/
	wg.Wait() // Ожидаем завершения всех горутин
}
