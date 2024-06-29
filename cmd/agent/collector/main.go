package collector

import (
	"fmt"
	"gmetrics/cmd/agent/collector/collection"
	"gmetrics/cmd/agent/collector/sender"
	"gmetrics/cmd/agent/env"
	"runtime"
	"time"
)

// StartCollect starts the process of collecting system metrics by launching a goroutine to execute the collectProcess function.
func StartCollect() {
	go collectProcess()
}

// collectProcess continuously collects system metrics by reading memory stats,
// collecting the stats using collection.Collection.Collect, and then sleeping
// for the duration defined by env.PollInterval.
func collectProcess() {
	fmt.Printf("Collect mtrics process starts. Period is %d mseconds\n", env.PollInterval)
	for {
		stats := runtime.MemStats{}
		runtime.ReadMemStats(&stats)
		collection.Collection.Collect(stats)
		time.Sleep(env.PollInterval)
	}
}

// StartSending starts the process of sending metrics by creating a new sender client and
// launching a goroutine to execute the periodicSender function.
func StartSending() *sender.Client {
	client := sender.New()
	fmt.Println("New sender client created")
	client.StartSender()

	return client
}
