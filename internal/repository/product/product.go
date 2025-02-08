package product

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Napat/golang-testcontainers-demo/internal/model"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	query := `
		INSERT INTO products (
			name, 
			description, 
			price, 
			sku, 
			stock
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.SKU,
		product.Stock,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	product := &model.Product{}
	query := `
		SELECT 
			id, 
			name, 
			description, 
			price, 
			sku, 
			stock,
			created_at,
			updated_at
		FROM products 
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SKU,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	query := `
		UPDATE products 
		SET 
			name = $1,
			description = $2,
			price = $3,
			sku = $4,
			stock = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING updated_at
	`
	result := r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.SKU,
		product.Stock,
		product.ID,
	)
	if err := result.Scan(&product.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return ErrProductNotFound
		}
		return err
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProductNotFound
	}
	return nil
}
