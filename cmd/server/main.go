package main

import "os"

func main() {
	if err := runServer(); err != nil {
		os.Exit(1)
	}
}
