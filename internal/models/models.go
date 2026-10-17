package models

import "time"

type User struct {
	ID        int64
	TgID      int64
	Username  string
	CreatedAt time.Time
}

type Item struct {
	ID                int64
	UserID            int64
	URL               string
	Title             string
	TargetPrice       float64
	LastPrice         *float64
	LastCurrency      *string
	LastCheckedAt     *time.Time
	LastNotifiedAt    *time.Time
	LastNotifiedPrice *float64
	CreatedAt         time.Time
}
