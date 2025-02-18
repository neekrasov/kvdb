package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/neekrasov/kvdb/internal/application"
	"github.com/neekrasov/kvdb/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()

	cfg, err := config.GetConfig("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	if err := application.New(&cfg).Start(ctx); err != nil {
		log.Fatal(err)
	}
}
