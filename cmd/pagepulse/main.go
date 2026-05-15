package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pagepulse/internal/app"
)

func main() {
	cfg, err := app.ParseConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	instance, err := app.New(cfg)
	if err != nil {
		log.Fatalf("startup error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := instance.Run(ctx); err != nil {
		log.Fatalf("runtime error: %v", err)
	}
}
