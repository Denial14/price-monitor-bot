package main

import (
	"context"
	"fmt"
	"os"
	"price-bot/internal/storage"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	_ = godotenv.Load()

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := storage.Connect(ctx, url)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	var version string
	err = pool.QueryRow(ctx, `SELECT version()`).Scan(&version)
	if err != nil {
		log.Fatalf("error getting version: %v", err)
	}
	fmt.Println("version: ", version)

	var usersCount, itemsCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&usersCount)
	if err != nil {
		log.Fatalf("error getting users count: %v", err)
	}

	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM items`).Scan(&itemsCount)
	if err != nil {
		log.Fatalf("error getting items count: %v", err)
	}
	fmt.Println("Users count: ", usersCount)
	fmt.Println("Items count: ", itemsCount)
}
