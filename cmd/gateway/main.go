package main

import "os"

func main() {
	if err := runGateway(); err != nil {
		os.Exit(1)
	}
}
