package main

import "os"

func main() {
	if err := runMigrate(); err != nil {
		os.Exit(1)
	}
}
