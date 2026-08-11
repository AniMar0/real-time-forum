package main

import (
	"real-time-forum/backend"
)

func main() {
	config := backend.LoadConfig()
	backend.MakeDataBaseAt(config.DatabasePath)
	var server backend.Server
	server.RunWithConfig(config)
}
