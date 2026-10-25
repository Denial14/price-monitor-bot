package storage

import (
	"context"
	"errors"
	"fmt"
	"price-bot/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) EnsureUser(ctx context.Context, tgID int64, username string) (*models.User, error) {
	var user models.User
	query := `
		INSERT INTO users (tg_id, username)
		VALUES ($1, $2)
		ON CONFLICT (tg_id)
		DO UPDATE SET username = EXCLUDED.username
		RETURNING id, tg_id, username, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := r.pool.QueryRow(ctx, query, tgID, username).Scan(&user.ID, &user.TgID, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("ensure user tg_id=%d: %w", tgID, err)
	}

	return &user, nil
}

func (r *UserRepo) GetByTgID(ctx context.Context, tgID int64) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, tg_id, username, created_at FROM users WHERE tg_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := r.pool.QueryRow(ctx, query, tgID).Scan(&user.ID, &user.TgID, &user.Username, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT id, tg_id, username, created_at FROM users WHERE id = $1`
	var user models.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(&user.ID, &user.TgID, &user.Username, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by user_id=%d: %w", userID, err)
	}
	return &user, nil
}
