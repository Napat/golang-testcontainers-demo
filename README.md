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
make pull-images

# "make pull-images" is equivalent to:
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
go test -v -short -cover -coverpkg=./internal/... ./test/integration/...

# Test all packages and generate HTML coverage report
go test -v -short -cover -coverprofile=coverage.out -coverpkg=./internal/... ./test/integration/...
go tool cover -html=coverage.out
```

Run specific tests

```sh
# MySQL integration tests - User Repository
go test -v -run Integration -short -cover -coverpkg=./internal/repository/repository_user/... ./test/integration/user/...

# PostgreSQL integration tests - Product Repository
go test -v -run Integration -short -cover -coverpkg=./internal/repository/repository_product/... ./test/integration/product/...

# Redis integration tests - Cache Repository
go test -v -run Integration -short -cover -coverpkg=./internal/repository/cache/... ./test/integration/cache/...

# Kafka integration tests - Event Producer Repository
go test -v -run Integration -short -cover -coverpkg=./internal/repository/repository_event/... ./test/integration/event/...

# ElasticSearch integration tests - Order Repository and API
go test -v -run Integration -short -cover -coverpkg=./internal/repository/repository_order/...,./internal/handler/... ./test/integration/order/...
```

## Test Coverage

This project includes comprehensive test coverage reporting tools and commands.

### Coverage Commands

```bash
# Generate overall coverage report
make coverage         # Shows total coverage percentage

# View detailed coverage by package
make coverage-by-package

# View coverage report in browser
make coverage-html
```

### Example Output

```bash
$ make coverage
ok      github.com/Napat/golang-testcontainers-demo/...    12.323s  coverage: 85.2% of statements
Total coverage: 85.2%

$ make coverage-by-package
github.com/Napat/golang-testcontainers-demo/pkg/model/user.go:31:      NewUser          100.0%
github.com/Napat/golang-testcontainers-demo/internal/repository/repository_user/user.go:45:    Create    90.0%
...
total:                                                                                85.2%
```

### Coverage Reports

- **HTML Report**: `coverage.html` - Visual representation of code coverage
- **Coverage File**: `coverage.out` - Raw coverage data
- **Package Report**: Detailed coverage by package and function

### Running Coverage Tests

1. Run all tests with coverage:

```bash
make coverage
```

2. View detailed package coverage:

```bash
make coverage-by-package
```

3. Open HTML coverage report:

```bash
make coverage-html
```

4. Run specific package coverage:

```bash
# MySQL repository coverage
make test-mysql

# PostgreSQL repository coverage
make test-postgres

# Redis cache coverage
make test-redis

# Kafka event coverage
make test-kafka

# Elasticsearch coverage
make test-elastic
```

## Database Initialization

The project uses the following directory structure for database initialization:

```ini
init/
├── mysql/
│   ├── 01-schema.sql
│   └── 02-data.sql (optional)
├── postgres/
│   ├── 01-schema.sql
│   └── 02-data.sql (optional)
└── elasticsearch/
    └── template.json
```

### Adding New Schemas

1. MySQL and PostgreSQL:

   - Add new `.sql` files in the respective directories
   - Files are executed in alphabetical order
   - Prefix files with numbers (e.g., `01-`, `02-`) to control execution order

2. Elasticsearch:

   - Add or modify templates in the `elasticsearch` directory
   - Templates are automatically applied when the container starts

Note: Changes to init scripts require container recreation to take effect:

```bash
docker compose down -v
docker compose up -d
```

## Request Tracing

This project uses OpenTelemetry with OTLP HTTP exporter to send traces to Jaeger for distributed tracing.

### Setup and Testing Steps

1. Start the Jaeger container and other services:

```bash
docker compose up -d
```

2. Run the application:

```bash
# Set configuration path
export CONFIG_PATH=configs/dev.yaml

# Start the application
go run ./cmd/api/main.go
```

3. The application should now be running on http://localhost:8080
4. Test the tracing by making some API calls:

```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "full_name": "John Doe",
    "password": "password123"
  }'

# Get user by ID
curl http://localhost:8080/users/1

# List all users
curl http://localhost:8080/users

# Send message to Kafka
curl -X POST http://localhost:8080/messages \
  -H 'Content-Type: application/json' \
  -d '{"content": "Hello World"}'

# Search documents in Elasticsearch
curl http://localhost:8080/search?q=example
```

5. View traces in Jaeger UI:

   - Open http://localhost:16686 in your browser
   - Select the service "testcontainers-demo"
   - Click "Find Traces" to see the trace details
   - You can see the complete request flow including:
      - HTTP endpoints
      - Database operations
      - Kafka messages
      - Elasticsearch queries

6. To stop the application:

   - Press Ctrl+C to stop the Go application
   - Run `docker compose down` to stop the containers

### Additional Tracing Configuration

The following information is captured in traces:

- HTTP request method and path
- Request duration
- Response status codes
- Service dependencies
- Database operations
- Cache operations
- Message broker interactions

### Trace Sampling

By default, all requests are sampled. Configure sampling in production by setting appropriate environment variables:

```sh
export OTEL_TRACES_SAMPLER=parentbased_traceidratio
export OTEL_TRACES_SAMPLER_ARG=0.1  # Sample 10% of requests
```

### Using Traces for Debugging

1. Find a trace in Jaeger UI by:

   - Service name
   - Operation name
   - Time range
   - Tags (HTTP method, status code)

2. Analyze trace details:

   - Request timeline
   - Service dependencies
   - Error messages
   - Operation durations

### Adding Custom Spans

Add custom spans to your code using the OpenTelemetry API:

```go
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// Add custom attributes
span.SetAttributes(attribute.String("key", "value"))
```

### Example API calls to test tracing:

```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Get user by ID
curl http://localhost:8080/users/1

# List all users
curl http://localhost:8080/users

# Create a product
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "sku": "PROD-001",
    "name": "Test Product",
    "description": "A test product",
    "price": 299.99,
    "stock": 100
  }'

# Create an order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "id": "order-1",
    "customer_id": "cust-1",
    "status": "pending",
    "total": 299.99,
    "payment_method": "credit_card",
    "items": [
      {
        "product_id": "prod-1",
        "product_name": "Test Product",
        "quantity": 1,
        "unit_price": 299.99,
        "subtotal": 299.99
      }
    ]
  }'

# Search orders with Elasticsearch query
curl -X GET http://localhost:8080/orders/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "match": {
        "customer_id": "cust-1"
      }
    }
  }'

# Send message to Kafka
curl -X POST http://localhost:8080/messages \
  -H 'Content-Type: application/json' \
  -d '{"content": "Hello World"}'

# Search documents in Elasticsearch
curl http://localhost:8080/search?q=example
```

After making these API calls, you can:

1. Open Jaeger UI at http://localhost:16686
2. Select the service "testcontainers-demo"
3. Click "Find Traces" to see the trace details
4. You can see the complete request flow including database operations, Kafka messages, and Elasticsearch queries

## Local Development

1. Start dependencies:

```bash
docker-compose up -d
```

2. Run the application:

```bash
# Development
CONFIG_PATH=configs/dev.yaml go run ./cmd/api/main.go

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

## API Documentation

### Swagger Documentation

1. Install required tools:

```bash
make dev-tools
```

2. Generate Swagger documentation:

```bash
make swagger    # Runs gen-swagger.sh to generate docs
```

3. View API documentation:

```bash
make serve-swagger   # Starts server with Swagger UI
```

Then open http://localhost:8080/swagger/index.html in your browser.

### Documentation Scripts

The project includes `scripts/gen-swagger.sh` which:

- Creates necessary directories
- Generates Swagger documentation
- Sets proper permissions
- Handles dependency parsing
- Supports internal package documentation

Usage:

```bash
./scripts/gen-swagger.sh    # Direct script execution
make swagger               # Using Makefile target
```

### Go Documentation

1. Start documentation server:

```bash
make serve-docs
```

2. View documentation:
   - Open browser: http://localhost:6060/pkg/github.com/Napat/golang-testcontainers-demo/
   - Navigate through packages

### Available Documentation

- API Endpoints

   - HTTP methods
   - Request/response schemas
   - Authentication requirements
   - Example requests

- Data Models

   - User
   - Product
   - Order
   - Error responses

### Adding Documentation

1. Add Swagger annotations to handlers:

```go
// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.User true "User object"
// @Success 201 {object} model.User
// @Router /users [post]
```

2. Generate updated documentation:

```bash
make swagger
```

## Performance Profiling

This project includes built-in performance profiling tools using Go's pprof package.

### Available Profiling Commands

```bash
# CPU Profiling
make prof-cpu           # Collect 30s CPU profile
make prof-flamegraph   # Generate CPU flamegraph

# Memory Profiling
make prof-mem          # Analyze heap allocations

# Goroutine Profiling
make prof-goroutine    # Analyze goroutine states

# Execution Tracing
make prof-trace        # Collect and analyze execution trace
```

### Using the Profiler

1. Start the application:

```bash
make run
```

2. Generate some load on your application
3. Run profiling commands:

```bash
# CPU profile with web UI
make prof-flamegraph

# Memory analysis
make prof-mem
```

4. Common profiling scenarios:

- CPU bottlenecks: `make prof-cpu`
- Memory leaks: `make prof-mem`
- Deadlocks/goroutine leaks: `make prof-goroutine`
- Concurrency issues: `make prof-trace`

### Profile Endpoints

The following endpoints are available:

- `/debug/pprof/` - Index page
- `/debug/pprof/profile` - CPU profile
- `/debug/pprof/heap` - Heap profile
- `/debug/pprof/goroutine` - Goroutine profile
- `/debug/pprof/trace` - Execution trace
- `/debug/pprof/block` - Block profile
- `/debug/pprof/threadcreate` - Thread creation

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

## Usage

### Users API

```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "full_name": "John Doe",
    "password": "password123"
  }'

# Get user by ID
curl http://localhost:8080/users/1

# List all users
curl http://localhost:8080/users
```

### Products API

```bash
# Create a new product
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "sku": "PROD-001",
    "name": "Test Product",
    "description": "A test product",
    "price": 299.99,
    "stock": 100
  }'

# Get all products
curl http://localhost:8080/products
```

### Orders API

```bash
# Create a new order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "id": "order-1",
    "customer_id": "cust-1",
    "status": "pending",
    "total": 299.99,
    "payment_method": "credit_card",
    "items": [
      {
        "product_id": "prod-1",
        "product_name": "Test Product",
        "quantity": 1,
        "unit_price": 299.99,
        "subtotal": 299.99
      }
    ]
  }'

# Search orders
curl -X GET http://localhost:8080/orders/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "match": {
        "customer_id": "cust-1"
      }
    }
  }'

# Simple search across all fields
curl "http://localhost:8080/search?q=Test%20Product"
```

## References

- [Testcontainers.com Getting started](https://testcontainers.com/getting-started/)
- [testcontainers modules](https://golang.testcontainers.org/modules/)
