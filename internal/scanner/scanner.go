package scanner

import (
	"context"
	"fmt"
	"sync"

	"price-bot/internal/models"
	"price-bot/internal/parser"
	"price-bot/internal/storage"

	"github.com/cenkalti/backoff/v4"
)

type Result struct {
	ItemID   int64
	NewPrice float64
	Currency string
	Error    error
}

type Scanner struct {
	items    *storage.ItemRepo
	limiter  *DomainLimiter
	registry *parser.Registry
	workers  int
	parseFn  func(ctx context.Context, p parser.Parser, url string) (parser.Price, error)
}

func New(items *storage.ItemRepo, limiter *DomainLimiter, registry *parser.Registry, workers int) *Scanner {
	return &Scanner{
		items:    items,
		limiter:  limiter,
		registry: registry,
		workers:  workers,
		parseFn:  defaultParseFN,
	}
}

func defaultParseFN(ctx context.Context, p parser.Parser, url string) (parser.Price, error) {
	var price parser.Price
	op := func() error {
		var err error
		price, err = p.Parse(ctx, url)
		return err
	}

	err := backoff.Retry(op, backoff.WithContext(backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3), ctx))
	return price, err
}

func (s *Scanner) ProcessItem(ctx context.Context, item models.Item) Result {
	if err := s.limiter.Wait(ctx, item.URL); err != nil {
		return Result{ItemID: item.ID, Error: err}
	}

	p := s.registry.For(item.URL)
	if p == nil {
		return Result{ItemID: item.ID, Error: fmt.Errorf("no parser for %s", item.URL)}
	}

	price, err := s.parseFn(ctx, p, item.URL) // ← используем parseFn
	if err != nil {
		return Result{ItemID: item.ID, Error: err}
	}

	if err := s.items.UpdatePrice(ctx, item.ID, price.Value, price.Currency); err != nil {
		return Result{ItemID: item.ID, Error: fmt.Errorf("update price: %w", err)}
	}
	return Result{ItemID: item.ID, NewPrice: price.Value, Currency: price.Currency}
}

func (s *Scanner) CheckAll(ctx context.Context, items []models.Item) []Result {
	tasks := make(chan models.Item, len(items))
	results := make(chan Result, len(items))

	for _, it := range items {
		tasks <- it
	}
	close(tasks)

	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for it := range tasks {
				res := s.ProcessItem(ctx, it)
				results <- res
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]Result, 0)
	for r := range results {
		out = append(out, r)
	}

	return out
}
