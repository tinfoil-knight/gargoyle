package main

import (
	"github.com/tinfoil-knight/gargoyle/internal/server"
)

func main() {
	server.Start("./config.json")
}
