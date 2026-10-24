package main

import (
	"context"
	"os/signal"
	"price-bot/internal/config"
	"price-bot/internal/parser"
	"price-bot/internal/scanner"
	"price-bot/internal/storage"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("conig: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := storage.Connect(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	items := storage.NewItemRepo(pool)
	registry := parser.NewRegistry()
	limiter := scanner.NewDomainLimiter(cfg.RateRPS, cfg.RateBurst)

	scan := scanner.New(items, limiter, registry, cfg.Workers)

	scan.RunOnce(ctx)

	ticker := time.NewTicker(cfg.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scanner: shutting down")
			return
		case <-ticker.C:
			scan.RunOnce(ctx)
		}
	}
}
