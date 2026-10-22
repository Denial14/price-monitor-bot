package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type BooksToScrape struct {
	client *http.Client
}

func NewBooksToScrape() *BooksToScrape {
	return &BooksToScrape{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (b *BooksToScrape) Match(url string) bool {
	return strings.Contains(url, "books.toscrape.com")
}

func (b *BooksToScrape) Parse(ctx context.Context, url string) (Price, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Price{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "PriceMonitorBot/1.0")

	resp, err := b.client.Do(req)
	if err != nil {
		return Price{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Price{}, fmt.Errorf("expected code 200, but got: %d", resp.StatusCode)
	}

	body := io.LimitReader(resp.Body, 5*1024*1024)
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return Price{}, fmt.Errorf("parse html: %w", err)
	}

	price, err := parseBooksToScrape(doc)
	return price, err
}

func parseBooksToScrape(doc *goquery.Document) (Price, error) {
	priceText := doc.Find("p.price_color").First().Text()
	if priceText == "" {
		return Price{}, fmt.Errorf("price element not found")
	}

	title := doc.Find("h1").First().Text()

	price, currency, err := parsePriceText(priceText)
	if err != nil {
		return Price{}, fmt.Errorf("parse price %q: %w", priceText, err)
	}

	return Price{
		Value:    price,
		Currency: currency,
		Title:    title,
	}, nil

}
