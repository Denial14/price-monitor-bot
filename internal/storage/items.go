package storage

import (
	"context"
	"errors"
	"fmt"
	"price-bot/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepo struct {
	pool *pgxpool.Pool
}

func NewItemRepo(pool *pgxpool.Pool) *ItemRepo {
	return &ItemRepo{pool: pool}
}

func (r *ItemRepo) Add(ctx context.Context, userID int64, url string, targetPrice float64, title string) (*models.Item, error) {
	query := `
		INSERT INTO items (user_id, url, target_price, title) 
		VALUES($1, $2, $3, $4) RETURNING id, user_id, target_price, url, title, created_at
	`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var item models.Item
	err := r.pool.QueryRow(ctx, query, userID, url, targetPrice, title).Scan(&item.ID, &item.UserID, &item.TargetPrice, &item.URL, &item.Title, &item.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgCodeUniqueViolation {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("add item: %w", err)
	}

	return &item, nil
}

func (r *ItemRepo) ListByUser(ctx context.Context, userID int64) ([]models.Item, error) {
	query := `
		SELECT id, user_id, url, title, target_price,
		last_price, last_currency, last_checked_at, last_notified_at,
		last_notified_price, created_at
		FROM items WHERE user_id = $1 ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("listByUser error getting by user_id=%d: %w", userID, err)
	}
	defer rows.Close()

	items := make([]models.Item, 0)
	for rows.Next() {
		var it models.Item
		err := rows.Scan(
			&it.ID, &it.UserID, &it.URL, &it.Title, &it.TargetPrice,
			&it.LastPrice, &it.LastCurrency, &it.LastCheckedAt,
			&it.LastNotifiedAt, &it.LastNotifiedPrice, &it.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}

func (r *ItemRepo) Remove(ctx context.Context, userID int64, itemID int64) error {
	query := `
		DELETE FROM items WHERE user_id = $1 AND id = $2 
	`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tag, err := r.pool.Exec(ctx, query, userID, itemID)
	if err != nil {
		return fmt.Errorf("remove item: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *ItemRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	query := `
		SELECT COUNT(*) FROM items WHERE user_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("getting items count user_id=%d: %w", userID, err)
	}
	return count, nil
}

func (r *ItemRepo) UpdatePrice(ctx context.Context, itemID int64, price float64, currency string) error {
	query := `
		UPDATE items
		SET last_price = $1, last_currency = $2, last_checked_at = now()
		WHERE id = $3
	`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := r.pool.Exec(ctx, query, price, currency, itemID)
	if err != nil {
		return fmt.Errorf("update price id=%d: %w", itemID, err)
	}
	return nil
}
