package scanner

import (
	"context"
	"net/url"
	"sync"

	"golang.org/x/time/rate"
)

type DomainLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rpc      float64
	burst    int
}

func NewDomainLimiter(rpc float64, burst int) *DomainLimiter {
	return &DomainLimiter{
		limiters: make(map[string]*rate.Limiter),
		rpc:      rpc,
		burst:    burst,
	}
}

func (l *DomainLimiter) Wait(ctx context.Context, url string) error {
	host := extractHost(url)

	l.mu.Lock()

	lim, ok := l.limiters[host]
	if !ok {
		lim = rate.NewLimiter(rate.Limit(l.rpc), l.burst)
		l.limiters[host] = lim
	}
	l.mu.Unlock()

	return lim.Wait(ctx)
}

func extractHost(rawURL string) string {
	url, err := url.Parse(rawURL)
	if err != nil || url.Host == "" {
		return ""
	}
	return url.Hostname()
}
