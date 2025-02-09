package cache

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Napat/golang-testcontainers-demo/internal/model"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/cache"
	"github.com/Napat/golang-testcontainers-demo/test/integration"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type CacheRepositoryTestSuite struct {
    integration.BaseTestSuite
    container testcontainers.Container
	ctx 	context.Context
    client    *redis.Client
    repo      *cache.Cache
}

// TestCacheRepository is the entry point for running the CacheRepositoryTestSuite.
// It sets up and executes all the tests defined in the CacheRepositoryTestSuite.
//
// Parameters:
//   - t: A testing.T object that manages the test state and supports logging.
//
// This function doesn't return any value. It uses the testify/suite package
// to run the entire test suite and reports the results through the testing.T object.
func TestCacheRepository(t *testing.T) {
    suite.Run(t, new(CacheRepositoryTestSuite))
}

// SetupSuite prepares the test environment for the CacheRepositoryTestSuite.
// 
// It sets up a Redis container using testcontainers, configures the container to
// log verbosely, uses a custom configuration file, and waits for the container to
// be ready to accept connections.
// 
// The method:
// - Initializes the base test suite
// - Creates a new context
// - Starts a Redis container with specified configuration
// - Establishes a connection to the Redis instance
// - Creates a new CacheRepository instance
// 
// The method doesn't take any parameters as it operates on the suite's fields.
// It doesn't return any values, but it populates the suite's fields with the necessary objects for testing.
func (s *CacheRepositoryTestSuite) SetupSuite() {
    s.BaseTestSuite.SetupSuite()
	s.ctx = context.Background()

    redisContainer, err := tcRedis.Run(s.ctx,
        "redis:6",
        tcRedis.WithSnapshotting(10, 1),
		tcRedis.WithLogLevel(tcRedis.LogLevelVerbose),
		tcRedis.WithConfigFile(filepath.Join("testdata", "redis6.conf")),
        testcontainers.WithWaitStrategy(
            wait.ForLog("Ready to accept connections").
                WithStartupTimeout(time.Minute),
        ),
    )
    s.Require().NoError(err)
    s.container = redisContainer

    // Get connection details
    host, err := redisContainer.Host(s.ctx)
    s.Require().NoError(err)
    port, err := redisContainer.MappedPort(s.ctx, "6379")
    s.Require().NoError(err)

    // Connect to Redis
    client := redis.NewClient(&redis.Options{
        Addr: fmt.Sprintf("%s:%s", host, port.Port()),
    })
    s.client = client

    // Verify connection
    s.Require().NoError(s.client.Ping(s.ctx).Err())

    // Initialize repository
    s.repo = cache.NewCacheRepository(client)
}

// TearDownSuite tears down the test environment for the CacheRepositoryTestSuite.
//
// The method:
// - Closes the Redis client connection
// - Cleans up the Redis container
//
// The method doesn't take any parameters as it operates on the suite's fields.
// It doesn't return any values.
func (s *CacheRepositoryTestSuite) TearDownSuite() {
    if s.client != nil {
        s.client.Close()
    }
    if s.container != nil {
        s.CleanupContainer(s.container)
    }
}

// TestSetAndGet tests the Set and Get methods of the CacheRepository.
//
// The test creates a sample User instance and uses Set to store it in Redis.
// Then, it uses Get to retrieve the same data from Redis and verifies that
// the retrieved instance matches the original one.
func (s *CacheRepositoryTestSuite) TestSetAndGet() {
    ctx := context.Background()
    
    // Test data
    user := &model.User{
        ID:       1,
        Username: "testuser",
        Email:    "test@example.com",
    }

    // Test Set
    err := s.repo.Set(ctx, "user:1", user, time.Minute)
    s.Require().NoError(err)

    // Test Get
    var fetchedUser model.User
    err = s.repo.Get(ctx, "user:1", &fetchedUser)
    s.Require().NoError(err)
    s.Equal(user.Username, fetchedUser.Username)
    s.Equal(user.Email, fetchedUser.Email)
}
