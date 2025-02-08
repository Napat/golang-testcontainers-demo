# Testcontainers Go Demo

This project demonstrates how to write integration tests using [Testcontainers for Go](https://golang.testcontainers.org/). It provides examples of testing different infrastructure components including databases, caching, and message queues.

## Overview

The project shows how to test applications that depend on:

- MySQL (User Repository)
- PostgreSQL (Product Repository)
- Redis (Cache Repository)
- Kafka (Event Producer)

## Prerequisites

- Go 1.20 or later
- Docker
- Docker Compose (optional)

## Preparation

Before running the tests, it's recommended to pull the required Docker images first to avoid timeouts during test execution:

```sh
# Pull required Docker images
docker pull mysql:8
docker pull postgres:14-alpine
docker pull redis:6
docker pull confluentinc/cp-kafka:7.8.0
```

## Project Structure

```sh
.
├── test/
│   └── integration/
│       ├── user/        # MySQL integration tests
│       ├── product/     # PostgreSQL integration tests
│       ├── cache/       # Redis integration tests
│       └── event/       # Kafka integration tests
```

## Running Tests

Run all integration tests:

```sh
go test -v ./test/integration/...
```

Run specific tests:

```sh
go test -v ./test/integration/user/...    # MySQL tests
go test -v ./test/integration/product/... # PostgreSQL tests
go test -v ./test/integration/cache/...   # Redis tests
go test -v ./test/integration/event/...   # Kafka tests
```

## Examples

### MySQL Example

```go
mysqlContainer, err := mysql.Run(ctx,
    "mysql:8",
    mysql.WithDatabase("testdb"),
    mysql.WithUsername("test"),
    mysql.WithPassword("test"),
)
```

### PostgreSQL Example

```go
postgresContainer, err := postgres.Run(ctx,
    "postgres:14-alpine",
    postgres.WithDatabase("testdb"),
    postgres.WithUsername("test"),
    postgres.WithPassword("test"),
)
```

### Redis Example

```go
redisContainer, err := tcRedis.Run(ctx,
    "redis:6",
    tcRedis.WithSnapshotting(10, 1),
)
```

### Kafka Example

```go
kafkaContainer, err := kafka.Run(ctx,
    "confluentinc/cp-kafka:7.8.0",
    kafka.WithClusterID("test-cluster"),
)
```

## Best Practices Demonstrated

1. Container Lifecycle Management
   - Proper setup and teardown
   - Resource cleanup
   - Timeout handling

2. Test Structure
   - Use of testify/suite
   - Shared base test suite
   - Isolated test environments

3. Configuration
   - Environment-specific settings
   - Clean test data management
   - Proper error handling

## Contributing

Feel free to contribute examples for other databases or infrastructure components!

## License

MIT License
