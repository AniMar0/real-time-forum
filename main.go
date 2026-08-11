package main

import (
	"context"
	"os"
	"os/signal"
	"real-time-forum/backend"
	"syscall"
	"time"
)

func main() {
	config := backend.LoadConfig()
	backend.MakeDataBaseAt(config.DatabasePath)
	var server backend.Server

	go server.RunWithConfig(config)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		os.Exit(1)
	}
}
