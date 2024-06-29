package main

import (
	"fmt"
	"gmetrics/cmd/agent/collector"
	"time"
)

func main() {
	fmt.Println("Agent is starting")
	collector.StartCollect()
	_ = collector.StartSending()

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
