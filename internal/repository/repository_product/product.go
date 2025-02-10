package repository_product

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Napat/golang-testcontainers-demo/pkg/model"
)

var ErrProductNotFound = errors.New("product not found")

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
            stock,
            version,
            created_at,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
        RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.SKU,
		product.Stock,
		1, // Initial version
	).Scan(&product.ID)

	return err
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
            updated_at,
            version
        FROM products
        WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SKU,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Version,
	)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]*model.Product, error) {
	query := `
        SELECT 
            id,
            name,
            description,
            price,
            sku,
            stock,
            created_at,
            updated_at,
            version
        FROM products
        ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		product := &model.Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.SKU,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.Version,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
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
            updated_at = NOW()
        WHERE id = $6`

	result, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.SKU,
		product.Stock,
		product.ID,
	)
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

func (r *ProductRepository) List(ctx context.Context) ([]*model.Product, error) {
	query := `
        SELECT 
            id,
            name,
            description,
            price,
            sku,
            stock,
            created_at,
            updated_at,
            version
        FROM products
        ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		product := &model.Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.SKU,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.Version,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}
