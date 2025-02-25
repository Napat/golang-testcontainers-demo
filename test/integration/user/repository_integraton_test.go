package user

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_user"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/Napat/golang-testcontainers-demo/pkg/testhelper"
	"github.com/Napat/golang-testcontainers-demo/test/integration"
	_ "github.com/go-sql-driver/mysql" // Add MySQL driver
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

type UserRepositoryTestSuite struct {
	integration.BaseTestSuite
	container testcontainers.Container
	db        *sql.DB
	repo      *repository_user.UserRepository
}

// TestIntegrationUserRepository runs the UserRepositoryTestSuite.
//
// This function sets up a test environment for testing the UserRepository type.
// It starts a MySQL container using testcontainers, sets up the database,
// and creates a new UserRepository instance for testing.
//
// The function doesn't take any parameters as it operates on the suite's fields.
// It doesn't return any values, but it populates the suite's fields with the necessary objects for testing.
func TestIntegrationUserRepository(t *testing.T) {
	testhelper.SkipIfShort(t)
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	ctx := context.Background()

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8",
		mysql.WithScripts(filepath.Join("testdata", "000002_alter_users_uuid.up.sql")),
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("test"),
		mysql.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server").
				WithStartupTimeout(time.Minute),
		),
	)
	s.Require().NoError(err)
	s.container = mysqlContainer

	// Get connection details
	host, err := mysqlContainer.Host(ctx)
	s.Require().NoError(err)
	port, err := mysqlContainer.MappedPort(ctx, "3306")
	s.Require().NoError(err)

	// Update connection string with proper parsing parameters
	dsn := fmt.Sprintf(
		"test:test@tcp(%s:%s)/testdb?parseTime=true&loc=UTC&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		host,
		port.Port(),
	)
	db, err := sql.Open("mysql", dsn)
	s.Require().NoError(err)
	s.db = db

	// Verify connection
	s.Require().NoError(s.db.Ping())

	// Create users table
	// _, err = db.Exec(`
	//     CREATE TABLE users (
	//         id BIGINT AUTO_INCREMENT PRIMARY KEY,
	//         username VARCHAR(255) NOT NULL,
	//         email VARCHAR(255) NOT NULL
	//     )
	// `)
	// s.Require().NoError(err)

	// Initialize repository
	s.repo = repository_user.NewUserRepository(db)
}

// TearDownSuite tears down the test environment for the UserRepositoryTestSuite.
//
// The method:
// - Closes the database connection
// - Cleans up the MySQL container
//
// The method doesn't take any parameters as it operates on the suite's fields.
// It doesn't return any values.

func (s *UserRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		s.CleanupContainer(s.container)
	}
}

// TestCreateAndGetUser tests the Create and GetByID methods of the
// UserRepository.
//
// The test creates a complete test user, and verifies that the Create method
// successfully creates the user and sets the ID field. It then verifies that the
// GetByID method retrieves the same user with matching fields.
func (s *UserRepositoryTestSuite) TestCreateAndGetUser() {
	ctx := context.Background()

	testUser := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",   // Make sure this is set
		Password: "password123", // Set password directly
		Status:   model.StatusActive,
	}

	err := s.repo.Create(ctx, testUser)
	s.Require().NoError(err)
	s.NotZero(testUser.ID)

	fetchedUser, err := s.repo.GetByID(ctx, testUser.ID)
	s.Require().NoError(err)
	s.Equal(testUser.Username, fetchedUser.Username)
	s.Equal(testUser.Email, fetchedUser.Email)
	s.Equal(testUser.FullName, fetchedUser.FullName)
	s.Equal(testUser.Password, fetchedUser.Password) // Changed from PasswordHash to Password
	s.Equal(testUser.Status, fetchedUser.Status)
	s.NotZero(fetchedUser.CreatedAt)
	s.NotZero(fetchedUser.UpdatedAt)
	s.Equal(1, fetchedUser.Version)
}
