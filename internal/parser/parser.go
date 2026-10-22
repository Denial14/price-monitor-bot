package parser

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Parser interface {
	Match(url string) bool
	Parse(ctx context.Context, url string) (Price, error)
}

type Price struct {
	Value    float64
	Currency string
	Title    string
}

type Registry struct {
	parsers []Parser
}

func NewRegistry() *Registry {
	return &Registry{
		parsers: []Parser{
			NewBooksToScrape(),
			NewOG(),
		},
	}
}

func (r *Registry) For(url string) Parser {
	for _, p := range r.parsers {
		if p.Match(url) {
			return p
		}
	}
	return nil
}

func parsePriceText(priceText string) (float64, string, error) {
	priceRe := regexp.MustCompile(`([₽$€£¥]|руб|RUB|USD|EUR|GBP)?\s*([\d\s.,]+)\s*([₽$€£¥]|руб|RUB|USD|EUR|GBP)?`)
	matches := priceRe.FindStringSubmatch(priceText)
	if len(matches) < 4 {
		return 0, "", fmt.Errorf("no price found in %q", priceText)
	}

	numberStr := strings.TrimSpace(matches[2])

	currency := matches[1]
	if currency == "" {
		currency = matches[3]
	}

	numberStr = strings.ReplaceAll(numberStr, ",", ".")
	numberStr = strings.ReplaceAll(numberStr, " ", "")
	numberStr = strings.TrimSpace(numberStr)

	price, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, "", fmt.Errorf("parse price: %w", err)
	}

	return price, currency, nil
}
