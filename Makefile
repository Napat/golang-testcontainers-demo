PROJECTNAME := $(shell basename "$(PWD)")
OS := $(shell uname -s | awk '{print tolower($$0)}')
GOARCH := amd64

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
BINARY_NAME=testcontainers-demo

# Build information
BUILD_TIME=$(shell date +%FT%T%z)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.buildTime=${BUILD_TIME} -X main.gitCommitSHA=${GIT_COMMIT}"

# Docker parameters
DOCKER_COMPOSE=docker compose

# Test parameters
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

#๒ all: Generate swagger, Run test, Build all components
.PHONY: all
all: swagger atest build

## version: Show version information
.PHONY: version
version:
	@echo "Build Time: ${BUILD_TIME}"
	@echo "Git Commit: ${GIT_COMMIT}"

## build: Build the application binary
.PHONY: build
build:
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/api

## run: Run the application
.PHONY: run
run:
	$(GORUN) ./cmd/api

## run-with-config: Run the application with dev config file
.PHONY: run-with-config
run-with-config:
	CONFIG_PATH=configs/dev.yaml $(GORUN) ./cmd/api

# Docker commands
## docker-up: Start all docker containers
.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) up -d

## docker-down: Stop all docker containers
.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

## docker-down-v: Stop and remove all docker containers and volumes
.PHONY: docker-down-v
docker-down-v:
	$(DOCKER_COMPOSE) down -v

## docker-logs: Show docker containers logs
.PHONY: docker-logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## docker-ps: List running docker containers
.PHONY: docker-ps
docker-ps:
	$(DOCKER_COMPOSE) ps

## test: Run unit tests
.PHONY: test
test:
	$(GOTEST) -cover -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep "total:" | awk '{print "\nTotal coverage: " $$3}'

## itest: Run integration test
.PHONY: itest
itest:
	$(GOTEST) -run Integration -short -cover -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep "total:" | awk '{print "\nTotal coverage: " $$3}'

## atest: Run both unit and integration test
.PHONY: atest
atest:
	$(GOTEST) -run Integration -short -cover -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep "total:" | awk '{print "\nTotal coverage: " $$3}'

## test-coverage: Run tests with coverage report
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -cover -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

## itest-mysql: Run MySQL integration tests
.PHONY: itest-mysql
itest-mysql:
	$(GOTEST) -v -run Integration -short -cover -coverpkg=./internal/repository/repository_user/... ./test/integration/user/...

## itest-postgres: Run PostgreSQL integration tests
.PHONY: itest-postgres
itest-postgres:
	$(GOTEST) -v -run Integration -short -cover -coverpkg=./internal/repository/repository_product/... ./test/integration/product/...

## itest-redis: Run Redis integration tests
.PHONY: itest-redis
itest-redis:
	$(GOTEST) -v -run Integration -short -cover -coverpkg=./internal/repository/cache/... ./test/integration/cache/...

## itest-kafka: Run Kafka integration tests
.PHONY: itest-kafka
itest-kafka:
	$(GOTEST) -v -run Integration -short -cover -coverpkg=./internal/repository/repository_event/... ./test/integration/event/...

## itest-elastic: Run Elasticsearch integration tests
.PHONY: itest-elastic
itest-elastic:
	$(GOTEST) -v -run Integration -short -cover -coverpkg=./internal/repository/repository_order/... ./test/integration/order/...

# Coverage commands
## coverage: Generate detailed coverage report
.PHONY: coverage
coverage:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic -coverpkg=./... ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep "total:" | awk '{print "Total coverage: " $$3}'
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

## coverage-by-package: Generate coverage report by package
.PHONY: coverage-by-package
coverage-by-package:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic -coverpkg=./... ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

## coverage-html: Open coverage report in browser
.PHONY: coverage-html
coverage-html: coverage
	open $(COVERAGE_HTML)

# Migration commands
## migrate: Run all database migrations
.PHONY: migrate
migrate:
	./scripts/migrate.sh --target all --command up

## migrate-down: Revert all database migrations
.PHONY: migrate-down
migrate-down:
	./scripts/migrate.sh --target all --command down

## migrate-mysql: Run MySQL migrations
.PHONY: migrate-mysql
migrate-mysql:
	./scripts/migrate.sh --target mysql --command up

## migrate-postgres: Run PostgreSQL migrations
.PHONY: migrate-postgres
migrate-postgres:
	./scripts/migrate.sh --target postgres --command up

## migrate-redis: Run Redis migrations
.PHONY: migrate-redis
migrate-redis:
	./scripts/migrate.sh --target redis --command up

## migrate-elastic: Run Elasticsearch migrations
.PHONY: migrate-elastic
migrate-elastic:
	./scripts/migrate.sh --target elasticsearch --command up

# Docker image preparation
## pull-images: Pull all required docker images
.PHONY: pull-images
pull-images:
	docker pull mysql:8
	docker pull postgres:14-alpine
	docker pull redis:6
	docker pull confluentinc/cp-kafka:7.8.0
	docker pull docker.elastic.co/elasticsearch/elasticsearch:8.17.1

# Clean up
## clean: Remove build artifacts and test coverage files
.PHONY: clean
clean:
	rm -f bin/$(BINARY_NAME)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

# Development helpers
## mock: Generate mock files for testing
.PHONY: mock
mock:
	mockgen -source=internal/repository/repository.go -destination=internal/repository/mock/mock_repository.go

## lint: Run golangci-lint
.PHONY: lint
lint:
	golangci-lint run

## tidy: Run go mod tidy
.PHONY: tidy
tidy:
	$(GOCMD) mod tidy

# Full development cycle
## dev-setup: Setup development environment
.PHONY: dev-setup
dev-setup: pull-images docker-up migrate

## dev-teardown: Teardown development environment
.PHONY: dev-teardown
dev-teardown: docker-down-v clean

## swagger: Generate Swagger documentation
.PHONY: swagger
swagger:
	chmod +x scripts/gen-swagger.sh
	./scripts/gen-swagger.sh

## serve-swagger: Run Swagger UI server
.PHONY: serve-swagger
serve-swagger: swagger
	$(GORUN) ./cmd/api

## serve-docs: Run godoc server
.PHONY: serve-docs
serve-docs:
	godoc -http=:6060

## dev-tools: Install development tools
.PHONY: dev-tools
dev-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/godoc@latest

## prof-cpu: Run CPU profiling
.PHONY: prof-cpu
prof-cpu:
	go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

## prof-mem: Run memory profiling
.PHONY: prof-mem
prof-mem:
	go tool pprof http://localhost:8080/debug/pprof/heap

## prof-goroutine: Run goroutine profiling
.PHONY: prof-goroutine
prof-goroutine:
	go tool pprof http://localhost:8080/debug/pprof/goroutine

## prof-trace: Generate execution trace
.PHONY: prof-trace
prof-trace:
	curl -o trace.out http://localhost:8080/debug/pprof/trace?seconds=30
	go tool trace trace.out

## prof-flamegraph: Generate CPU flamegraph
.PHONY: prof-flamegraph
prof-flamegraph:
	go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile?seconds=30

## help: Display help messages
.PHONY: help
# all: help
help: Makefile
	@echo
	@echo " Project: ["$(PROJECTNAME)"]"
	@echo " Please choose a command"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
