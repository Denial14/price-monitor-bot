package storage

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrItemNotFound = errors.New("item not found")
	ErrDuplicate    = errors.New("duplicate entry")
)

const (
	pgCodeUniqueViolation     = "23505"
	pgCodeForeignKeyViolation = "23503"
)
