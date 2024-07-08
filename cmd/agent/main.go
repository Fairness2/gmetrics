package main

import (
	"fmt"
	"gmetrics/cmd/agent/collector"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/config"
	"time"
)

func main() {
	fmt.Println("Agent is starting")
	// Устанавливаем настройки
	config.Parse()
	fmt.Print(config.PrintConfig())

	collector.StartCollect()                          // Запускаем сборку данных
	_ = collector.StartSending(collection.Collection) // Запускаем отправку данных

	time.Sleep(30 * time.Second)

	/*for {
		fmt.Println("Agent is running. Print C to finish agent")
		var command string
		_, err := fmt.Fscan(os.Stdin, &command)
		if err != nil {
			fmt.Println(err)
		}
		if command == "C" {
			break
		}
	}*/
}
