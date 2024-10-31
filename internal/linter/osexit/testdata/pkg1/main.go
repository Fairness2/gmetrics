package main

import "os"

func main() {
	os.Exit(1) // want "os exit is not prohibited"
}
