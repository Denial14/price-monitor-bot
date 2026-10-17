package main

import (
	"fmt"
	"log"
	"price-bot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Printf("%+v\n", cfg)
	fmt.Printf("%s", cfg.DatabaseURL())
}
