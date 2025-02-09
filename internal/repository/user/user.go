package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Napat/golang-testcontainers-demo/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
    query := `
        INSERT INTO users (
            username, 
            email, 
            full_name,
            password, 
            status
        ) VALUES (?, ?, ?, ?, ?)
    `
    result, err := r.db.ExecContext(ctx, query,
        user.Username,
        user.Email,
        user.FullName,
        user.Password,
        user.Status,
    )
    if err != nil {
        return err
    }
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    user.ID = id
    return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
    user := &model.User{}
    query := `
        SELECT 
            id, 
            username, 
            email, 
            full_name,
            password,
            status, 
            created_at, 
            updated_at,
            version
        FROM users 
        WHERE id = ?
    `
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.FullName,
        &user.Password,
        &user.Status,
        &user.CreatedAt,
        &user.UpdatedAt,
        &user.Version,
    )
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, err
    }
    return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
    query := `
        UPDATE users 
        SET 
            username = ?,
            email = ?,
            full_name = ?,
            status = ?
        WHERE id = ?
    `
    result, err := r.db.ExecContext(ctx, query,
        user.Username,
        user.Email,
        user.FullName,
        user.Status,
        user.ID,
    )
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return ErrUserNotFound
    }
    return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
    result, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return ErrUserNotFound
    }
    return nil
}
