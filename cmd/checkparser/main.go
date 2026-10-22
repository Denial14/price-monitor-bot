package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"price-bot/internal/parser"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: checkparser <url>")
	}
	rawURL := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry := parser.NewRegistry()
	p := registry.For(rawURL)
	if p == nil {
		log.Fatalf("no parser for %s", rawURL)
	}

	price, err := p.Parse(ctx, rawURL)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}

	fmt.Printf("title:    %s\n", price.Title)
	fmt.Printf("value:    %.2f\n", price.Value)
	fmt.Printf("currency: %s\n", price.Currency)
}
