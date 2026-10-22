package helpers

import (
	"fmt"
	"net/url"
	"price-bot/internal/models"
	"strconv"
	"strings"
	"unicode/utf8"
)

func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https, but got: %s", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("host is empty")
	}
	return nil
}

func ValidatePrice(rawPrice string) (float64, error) {
	replacePrice := strings.ReplaceAll(rawPrice, ",", ".")
	price, err := strconv.ParseFloat(replacePrice, 64)
	if err != nil {
		return 0, fmt.Errorf("price must be a number, but got %q: %w", rawPrice, err)
	}
	if price <= 0 {
		return 0, fmt.Errorf("price must be positive, but got: %v", price)
	}
	return price, nil
}

func ValidateTitle(title string) string {
	if utf8.RuneCountInString(title) > 200 {
		newTitle := []rune(title)[:200]
		return string(newTitle)
	}
	return title
}

func FormatItem(item models.Item) string {
	title := item.Title
	if title == "" {
		title = item.URL
	}

	runes := []rune(title)
	if len(runes) > 50 {
		title = string(runes[:47]) + "..."
	}
	return fmt.Sprintf("[%d] %s - %.2f", item.ID, title, item.TargetPrice)
}
