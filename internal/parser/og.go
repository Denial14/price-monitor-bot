package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type OG struct {
	client *http.Client
}

func NewOG() *OG {
	return &OG{client: &http.Client{Timeout: 15 * time.Second}}
}

func (o *OG) Match(_ string) bool {
	return true
}

func (o *OG) Parse(ctx context.Context, url string) (Price, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Price{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "PriceMonitorBot/1.0")

	resp, err := o.client.Do(req)
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

	price, ok := parseOG(doc)
	if !ok {
		return Price{}, fmt.Errorf("parse OG")
	}
	return price, nil
}

func parseOG(doc *goquery.Document) (Price, bool) {
	amount, ok := doc.Find(`meta[property="product:price:amount"]`).First().Attr("content")
	if !ok {
		return Price{}, false
	}

	currency, _ := doc.Find(`meta[property="product:price:currency"]`).First().Attr("content")
	title, _ := doc.Find(`meta[property="og:title"]`).First().Attr("content")

	price, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return Price{}, false
	}

	return Price{
		Value:    price,
		Currency: currency,
		Title:    title,
	}, true
}
