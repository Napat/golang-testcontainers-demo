package repository_user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Napat/golang-testcontainers-demo/pkg/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if user.Status == "" {
		user.Status = model.StatusActive
	}

	query := `
        INSERT INTO users (
            username,
            email,
            full_name,
            password,
            status,
            version,
            created_at,
            updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`

	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.FullName,
		user.Password,
		string(user.Status), // Convert UserStatus to string
		1,                   // Initial version
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
        WHERE id = ?`

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

func (r *UserRepository) GetAll(ctx context.Context) ([]*model.User, error) {
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
        ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
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
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
        UPDATE users
        SET
            username = ?,
            email = ?,
            full_name = ?,
            status = ?
        WHERE id = ?`

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
