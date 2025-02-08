package integration

import (
	"context"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type BaseTestSuite struct {
    suite.Suite
    ctx context.Context
}

// SetupSuite sets up the test environment for the BaseTestSuite.
//
// This method:
// - Creates a context with a 5-minute timeout for the test suite
// - Assigns the context to the suite's ctx field
//
// It doesn't take any parameters as it operates on the suite's fields.
// It does not return any values.

func (s *BaseTestSuite) SetupSuite() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    s.ctx = ctx
}

// CleanupContainer terminates the given testcontainers.Container.
//
// This method:
// - Creates a context with a 30-second timeout for container termination
// - Attempts to terminate the container using the context
// - Logs a message if the container termination fails
//
// It takes the following parameter:
// - container: The testcontainers.Container to be terminated
//
// It doesn't return any values.

func (s *BaseTestSuite) CleanupContainer(container testcontainers.Container) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := container.Terminate(ctx); err != nil {
        s.T().Logf("failed to terminate container: %v", err)
    }
}
