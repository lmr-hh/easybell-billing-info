package main

import (
	"os"
)

func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
