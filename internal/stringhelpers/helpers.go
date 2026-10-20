package stringhelpers

import (
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

func ValidateURL(rawURL string) error {
	url, err := url.Parse(rawURL)
	if err != nil || (url.Scheme != "http" && url.Scheme != "https" || url.Host == "") {
		return err
	}
	return nil
}

func ValidatePrice(rawPrice string) (float64, error) {
	replacePrice := strings.ReplaceAll(rawPrice, ",", ".")
	price, err := strconv.ParseFloat(replacePrice, 64)
	if err != nil || price <= 0 {
		return 0, err
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
