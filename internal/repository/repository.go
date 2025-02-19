package repository

import (
    "context"
    "errors"
    "github.com/google/uuid"
)

type Repository[T any] interface {
    Create(ctx context.Context, entity *T) error
    GetByID(ctx context.Context, id uuid.UUID) (*T, error)
    Update(ctx context.Context, entity *T) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// Common errors
var (
    ErrNotFound = errors.New("entity not found")
    ErrInvalid  = errors.New("invalid entity")
)
