package main

import (
	"os"

	"github.com/tinfoil-knight/gargoyle/internal/server"
)

func main() {
	var configFile string
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	if configFile == "" {
		configFile = "./config.json"
	}
	server.Start(configFile)
}
