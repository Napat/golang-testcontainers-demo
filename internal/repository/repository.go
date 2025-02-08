package repository

import (
	"context"
	"errors"
)

type Repository[T any] interface {
    Create(ctx context.Context, entity *T) error
    GetByID(ctx context.Context, id int64) (*T, error)
    Update(ctx context.Context, entity *T) error
    Delete(ctx context.Context, id int64) error
}

// Common errors
var (
    ErrNotFound = errors.New("entity not found")
    ErrInvalid  = errors.New("invalid entity")
)
