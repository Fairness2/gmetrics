package main

import (
	"fmt"
	"gmetrics/cmd/agent/collector"
	"os"
)

func main() {
	fmt.Println("Agent is starting")
	collector.StartCollect()
	_ = collector.StartSending()

	for {
		fmt.Println("Agent is running. Print C to finish agent")
		var command string
		_, err := fmt.Fscan(os.Stdin, &command)
		if err != nil {
			fmt.Println(err)
		}
		if command == "C" {
			break
		}
	}
}
