package product

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Napat/golang-testcontainers-demo/internal/model"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/product"
	"github.com/Napat/golang-testcontainers-demo/test/integration"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ProductRepositoryTestSuite struct {
	integration.BaseTestSuite
	container testcontainers.Container
    ctx context.Context
	db        *sql.DB
	repo      *product.ProductRepository
}

// TestProductRepository is a test suite for the ProductRepository type.
// It uses Testcontainers to spin up a PostgreSQL database container,
// applies the necessary migration scripts, and tests the CRUD operations
// of the ProductRepository.
func TestProductRepository(t *testing.T) {
    suite.Run(t, new(ProductRepositoryTestSuite))
}

// SetupSuite prepares the test environment for the ProductRepositoryTestSuite.
// It initializes a PostgreSQL container using testcontainers, sets up the database,
// and creates a new ProductRepository instance for testing.
//
// This method:
// - Initializes the base test suite
// - Creates a new context
// - Starts a PostgreSQL container with specified configuration
// - Establishes a connection to the database
// - Creates a new ProductRepository instance
//
// The method doesn't take any parameters as it operates on the suite's fields.
// It doesn't return any values, but it populates the suite's fields with the necessary objects for testing.
func (s *ProductRepositoryTestSuite) SetupSuite() {
    s.BaseTestSuite.SetupSuite()
    s.ctx = context.Background()

    postgresContainer, err := postgres.Run(s.ctx,
        "postgres:14-alpine",
        postgres.WithInitScripts(filepath.Join("testdata", "000001_create_products_table.up.sql")),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).WithStartupTimeout(5*time.Second),
        ),
    )
    s.Require().NoError(err)
    s.container = postgresContainer

    host, err := postgresContainer.Host(s.ctx)
    s.Require().NoError(err)

    mappedPort, err := postgresContainer.MappedPort(s.ctx, "5432/tcp")
    s.Require().NoError(err)

    dsn := fmt.Sprintf(
        "postgres://test:test@%s:%s/testdb?sslmode=disable",
        host,
        mappedPort.Port(),
    )

    // Add retry logic for database connection
    var db *sql.DB
    for i := 0; i < 5; i++ {
        db, err = sql.Open("postgres", dsn)
        if err != nil {
            time.Sleep(time.Second)
            continue
        }
        if err = db.Ping(); err == nil {
            break
        }
        time.Sleep(time.Second)
    }
    s.Require().NoError(err)
    s.db = db
    s.repo = product.NewProductRepository(db)
}

// TearDownSuite tears down the test environment for the ProductRepositoryTestSuite.
//
// The method:
// - Closes the database connection
// - Cleans up the PostgreSQL container
//
// The method doesn't take any parameters as it operates on the suite's fields.
// It doesn't return any values.
func (s *ProductRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		s.CleanupContainer(s.container)
	}
}

// TestCreateAndGetProduct tests the Create and GetByID methods of the
// ProductRepository.
//
// The method:
// - Creates a test product using the Create method
// - Verifies that the product ID is not zero
// - Retrieves the created product using the GetByID method
// - Verifies that the retrieved product matches the original test product
func (s *ProductRepositoryTestSuite) TestCreateAndGetProduct() {
	ctx := context.Background()

	testProduct := &model.Product{
		Name:        "Test Product",
		Description: "This is a test product",
		Price:       29.99,
		SKU:         "TEST-001",
		Stock:       50,
	}

	err := s.repo.Create(ctx, testProduct)
	s.Require().NoError(err)
	s.NotZero(testProduct.ID)

	fetchedProduct, err := s.repo.GetByID(ctx, testProduct.ID)
	s.Require().NoError(err)
	s.Equal(testProduct.Name, fetchedProduct.Name)
	s.Equal(testProduct.SKU, fetchedProduct.SKU)
	s.Equal(testProduct.Price, fetchedProduct.Price)
	s.Equal(testProduct.Stock, fetchedProduct.Stock)
}
