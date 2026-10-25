package main

import (
	"context"
	"os/signal"
	"price-bot/internal/bot"
	"price-bot/internal/config"
	"price-bot/internal/parser"
	"price-bot/internal/storage"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	token := cfg.BotToken
	if token == "" {
		log.Fatal("token is empty")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := storage.Connect(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Errorf("connect: %v", err)
	}
	defer pool.Close()

	registry := parser.NewRegistry()
	b, err := bot.New(cfg.BotToken, pool, registry, cfg.PollInterval)
	if err != nil {
		log.Errorf("bot: %v", err)
	}

	if err := b.Run(ctx); err != nil {
		log.Errorf("run bot: %v", err)
	}
}
