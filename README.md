# Testcontainers Go Demo

This project demonstrates how to write integration tests using [Testcontainers for Go](https://golang.testcontainers.org/). It provides examples of testing different infrastructure components including databases, caching, and message queues.

## Overview

The project shows how to test applications that depend on:

- MySQL (User Repository)
- PostgreSQL (Product Repository)
- Redis (Cache Repository)
- Kafka (Event Producer)
- Elasticsearch (Order Repository)

## Prerequisites

- Go 1.20 or later
- Docker
- Docker Compose (optional)

## Preparation

Before running the tests, it's recommended to pull the required Docker images first to avoid timeouts during test execution.

```sh
# Pull required Docker images
docker pull mysql:8
docker pull postgres:14-alpine
docker pull redis:6
docker pull confluentinc/cp-kafka:7.8.0
docker pull docker.elastic.co/elasticsearch/elasticsearch:8.17.1
```

## Project Structure

```sh
.
├── test/
│   └── integration/
│       ├── user/        # MySQL integration tests
│       ├── product/     # PostgreSQL integration tests
│       ├── cache/       # Redis integration tests
│       ├── event/       # Kafka integration tests
│       └── order/       # Elasticsearch integration tests
```

## Running Tests

Run all integration tests with coverage:

```sh
# Test all packages and show coverage
go test -v -cover -coverpkg=./internal/... ./test/integration/...

# Generate HTML coverage report
go test -v -cover -coverprofile=coverage.out -coverpkg=./internal/... ./test/integration/...
go tool cover -html=coverage.out
```

Run specific tests:

```sh
# MySQL tests - User Repository
go test -v -cover -coverpkg=./internal/repository/user/... ./test/integration/user/...

# PostgreSQL tests - Product Repository
go test -v -cover -coverpkg=./internal/repository/product/... ./test/integration/product/...

# Redis tests - Cache Repository
go test -v -cover -coverpkg=./internal/repository/cache/... ./test/integration/cache/...

# Kafka tests - Event Producer Repository
go test -v -cover -coverpkg=./internal/repository/event/... ./test/integration/event/...

# ElasticSearch tests - Order Repository and API
go test -v -cover -coverpkg=./internal/repository/order/...,./internal/handler/... ./test/integration/order/...
```

## Local Development

1. Start dependencies:

```bash
docker-compose up -d
```

2. Run the application:

```bash
# Development
CONFIG_PATH=configs/dev.yaml go run main.go

# Test
CONFIG_PATH=configs/test.yaml go test -v ./test/integration/...
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

### Elasticsearch Example

```go
req := testcontainers.ContainerRequest{
    Image:        "docker.elastic.co/elasticsearch/elasticsearch:8.17.1",
    ExposedPorts: []string{"9200/tcp"},
    Env: map[string]string{
        "discovery.type":         "single-node",
        "xpack.security.enabled": "false",
        "ES_JAVA_OPTS":          "-Xms512m -Xmx512m",
    },
    WaitingFor: wait.ForHTTP("/").WithPort("9200"),
    HostConfigModifier: func(hostConfig *container.HostConfig) {
        hostConfig.Resources = container.Resources{
            Memory:            2 * 1024 * 1024 * 1024, // 2GB limit
            MemoryReservation: 1024 * 1024 * 1024,     // 1GB reservation
        },
    },
}
```

The Elasticsearch example demonstrates:

- Setting up single-node cluster for testing
- Memory optimization for containers
- Index template management
- Data seeding for tests
- Search query testing

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
