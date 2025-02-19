package repository_product

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Napat/golang-testcontainers-demo/pkg/metrics"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository struct {
	db      *sql.DB
	metrics *metrics.DatabaseMetrics
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db:      db,
		metrics: metrics.NewDatabaseMetrics("product"),
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("create", "products").Observe(time.Since(timer).Seconds())
	}()

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

	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("create", "products", "error").Inc()
		return err
	}

	r.metrics.QueriesTotal.WithLabelValues("create", "products", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("postgres").Set(float64(r.db.Stats().OpenConnections))
	return err
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *model.Product) error {
	query := `INSERT INTO products (name, price, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id`
	return r.db.QueryRowContext(ctx, query, product.Name, product.Price).Scan(&product.ID)
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("get", "products").Observe(time.Since(timer).Seconds())
	}()

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
		r.metrics.QueriesTotal.WithLabelValues("get", "products", "error").Inc()
		return nil, ErrProductNotFound
	}
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("get", "products", "error").Inc()
		return nil, err
	}

	r.metrics.QueriesTotal.WithLabelValues("get", "products", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("postgres").Set(float64(r.db.Stats().OpenConnections))
	return product, nil
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]*model.Product, error) {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("list", "products").Observe(time.Since(timer).Seconds())
	}()

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
		r.metrics.QueriesTotal.WithLabelValues("list", "products", "error").Inc()
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
			r.metrics.QueriesTotal.WithLabelValues("list", "products", "error").Inc()
			return nil, err
		}
		products = append(products, product)
	}

	r.metrics.QueriesTotal.WithLabelValues("list", "products", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("postgres").Set(float64(r.db.Stats().OpenConnections))
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("update", "products").Observe(time.Since(timer).Seconds())
	}()

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
		r.metrics.QueriesTotal.WithLabelValues("update", "products", "error").Inc()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("update", "products", "error").Inc()
		return err
	}
	if rows == 0 {
		r.metrics.QueriesTotal.WithLabelValues("update", "products", "error").Inc()
		return ErrProductNotFound
	}

	r.metrics.QueriesTotal.WithLabelValues("update", "products", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("postgres").Set(float64(r.db.Stats().OpenConnections))
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("delete", "products").Observe(time.Since(timer).Seconds())
	}()

	result, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("delete", "products", "error").Inc()
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.metrics.QueriesTotal.WithLabelValues("delete", "products", "error").Inc()
		return err
	}
	if rows == 0 {
		r.metrics.QueriesTotal.WithLabelValues("delete", "products", "error").Inc()
		return ErrProductNotFound
	}

	r.metrics.QueriesTotal.WithLabelValues("delete", "products", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("postgres").Set(float64(r.db.Stats().OpenConnections))
	return nil
}

func (r *ProductRepository) List(ctx context.Context) ([]*model.Product, error) {
	timer := time.Now()
	defer func() {
		r.metrics.QueryDuration.WithLabelValues("list", "products").Observe(time.Since(timer).Seconds())
	}()

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
		r.metrics.QueriesTotal.WithLabelValues("list", "products", "error").Inc()
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
			r.metrics.QueriesTotal.WithLabelValues("list", "products", "error").Inc()
			return nil, err
		}
		products = append(products, product)
	}

	r.metrics.QueriesTotal.WithLabelValues("list", "products", "success").Inc()
	r.metrics.ConnectionsOpen.WithLabelValues("postgres").Set(float64(r.db.Stats().OpenConnections))
	return products, nil
}
