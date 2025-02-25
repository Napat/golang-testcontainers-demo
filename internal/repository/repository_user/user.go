package repository_user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Napat/golang-testcontainers-demo/pkg/metrics"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db      *sql.DB
	metrics *metrics.DatabaseMetrics
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db:      db,
		metrics: metrics.NewDatabaseMetrics("user"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("create", "users").Observe(time.Since(timer).Seconds())
	}()

	if user.Status == "" {
		user.Status = model.StatusActive
	}

	// Generate UUIDv7 if not set
	if user.ID == uuid.Nil {
		user.ID = uuid.Must(uuid.NewV7())
	}

	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Defer a rollback in case anything fails
	defer tx.Rollback()

	// Check for existing username and email within the transaction
	var exists int
	err = tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE username = ? OR email = ? FOR UPDATE",
		user.Username, user.Email).Scan(&exists)
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("create", "users", "error").Inc()
		return fmt.Errorf("failed to check username/email existence: %w", err)
	}
	if exists > 0 {
		// Check which one is duplicate
		var existingUser model.User
		err = tx.QueryRowContext(ctx,
			"SELECT username, email FROM users WHERE username = ? OR email = ?",
			user.Username, user.Email).Scan(&existingUser.Username, &existingUser.Email)
		if err != nil && err != sql.ErrNoRows {
			r.metrics.QueriesTotal.WithLabelValues("create", "users", "error").Inc()
			return err
		}
		if existingUser.Username == user.Username {
			r.metrics.QueriesTotal.WithLabelValues("create", "users", "error").Inc()
			return fmt.Errorf("Error 1062 (23000): Duplicate entry '%s' for key 'users.username'", user.Username)
		}
		r.metrics.QueriesTotal.WithLabelValues("create", "users", "error").Inc()
		return fmt.Errorf("Error 1062 (23000): Duplicate entry '%s' for key 'users.email'", user.Email)
	}

	query := `
        INSERT INTO users (
            id,
            username,
            email,
            full_name,
            password,
            status,
            version,
            created_at,
            updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`

	_, err = tx.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.FullName,
		user.Password,
		string(user.Status),
		1,
	)
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("create", "users", "error").Inc()
		return err
	}

	// If everything went well, commit the transaction
	if err = tx.Commit(); err != nil {
		r.metrics.QueriesTotal.WithLabelValues("create", "users", "error").Inc()
		return err
	}

	r.metrics.QueriesTotal.WithLabelValues("create", "users", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("mysql").Set(float64(r.db.Stats().OpenConnections))

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("get_by_id", "users").Observe(time.Since(timer).Seconds())
	}()

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
		r.metrics.QueriesTotal.WithLabelValues("get_by_id", "users", "error").Inc()
		return nil, ErrUserNotFound
	}
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("get_by_id", "users", "error").Inc()
		return nil, err
	}

	r.metrics.QueriesTotal.WithLabelValues("get_by_id", "users", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("mysql").Set(float64(r.db.Stats().OpenConnections))

	return user, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*model.User, error) {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("get_all", "users").Observe(time.Since(timer).Seconds())
	}()

	query := `
        SELECT
            ROW_NUMBER() OVER (ORDER BY id) as row_num,
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
		r.metrics.QueriesTotal.WithLabelValues("get_all", "users", "error").Inc()
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.RowNumber,
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
			r.metrics.QueriesTotal.WithLabelValues("get_all", "users", "error").Inc()
			return nil, err
		}
		users = append(users, user)
	}

	r.metrics.QueriesTotal.WithLabelValues("get_all", "users", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("mysql").Set(float64(r.db.Stats().OpenConnections))

	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("update", "users").Observe(time.Since(timer).Seconds())
	}()

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
		r.metrics.QueriesTotal.WithLabelValues("update", "users", "error").Inc()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("update", "users", "error").Inc()
		return err
	}
	if rows == 0 {
		r.metrics.QueriesTotal.WithLabelValues("update", "users", "error").Inc()
		return ErrUserNotFound
	}

	r.metrics.QueriesTotal.WithLabelValues("update", "users", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("mysql").Set(float64(r.db.Stats().OpenConnections))

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("delete", "users").Observe(time.Since(timer).Seconds())
	}()

	result, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("delete", "users", "error").Inc()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("delete", "users", "error").Inc()
		return err
	}
	if rows == 0 {
		r.metrics.QueriesTotal.WithLabelValues("delete", "users", "error").Inc()
		return ErrUserNotFound
	}

	r.metrics.QueriesTotal.WithLabelValues("delete", "users", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("mysql").Set(float64(r.db.Stats().OpenConnections))

	return nil
}
