# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
BINARY_NAME=testcontainers-demo

# Docker parameters
DOCKER_COMPOSE=docker compose

# Test parameters
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

.PHONY: all
all: swagger test build

.PHONY: build
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/api

.PHONY: run
run:
	$(GORUN) ./cmd/api

.PHONY: run-with-config
run-with-config:
	CONFIG_PATH=configs/dev.yaml $(GORUN) ./cmd/api

# Docker commands
.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) up -d

.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

.PHONY: docker-down-v
docker-down-v:
	$(DOCKER_COMPOSE) down -v

.PHONY: docker-logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f

.PHONY: docker-ps
docker-ps:
	$(DOCKER_COMPOSE) ps

# Test commands
.PHONY: test
test:
	$(GOTEST) -cover -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep "total:" | awk '{print "\nTotal coverage: " $$3}'

.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -cover -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

.PHONY: test-mysql
test-mysql:
	$(GOTEST) -v -cover -coverpkg=./internal/repository/repository_user/... ./test/integration/user/...

.PHONY: test-postgres
test-postgres:
	$(GOTEST) -v -cover -coverpkg=./internal/repository/repository_product/... ./test/integration/product/...

.PHONY: test-redis
test-redis:
	$(GOTEST) -v -cover -coverpkg=./internal/repository/cache/... ./test/integration/cache/...

.PHONY: test-kafka
test-kafka:
	$(GOTEST) -v -cover -coverpkg=./internal/repository/repository_event/... ./test/integration/event/...

.PHONY: test-elastic
test-elastic:
	$(GOTEST) -v -cover -coverpkg=./internal/repository/repository_order/... ./test/integration/order/...

# Coverage commands
.PHONY: coverage
coverage:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic -coverpkg=./... ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep "total:" | awk '{print "Total coverage: " $$3}'
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

.PHONY: coverage-by-package
coverage-by-package:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic -coverpkg=./... ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

.PHONY: coverage-html
coverage-html: coverage
	open $(COVERAGE_HTML)

# Migration commands
.PHONY: migrate
migrate:
	./scripts/migrate.sh --target all --command up

.PHONY: migrate-down
migrate-down:
	./scripts/migrate.sh --target all --command down

.PHONY: migrate-mysql
migrate-mysql:
	./scripts/migrate.sh --target mysql --command up

.PHONY: migrate-postgres
migrate-postgres:
	./scripts/migrate.sh --target postgres --command up

.PHONY: migrate-redis
migrate-redis:
	./scripts/migrate.sh --target redis --command up

.PHONY: migrate-elastic
migrate-elastic:
	./scripts/migrate.sh --target elasticsearch --command up

# Docker image preparation
.PHONY: pull-images
pull-images:
	docker pull mysql:8
	docker pull postgres:14-alpine
	docker pull redis:6
	docker pull confluentinc/cp-kafka:7.8.0
	docker pull docker.elastic.co/elasticsearch/elasticsearch:8.17.1

# Clean up
.PHONY: clean
clean:
	rm -f bin/$(BINARY_NAME)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

# Development helpers
.PHONY: mock
mock:
	mockgen -source=internal/repository/repository.go -destination=internal/repository/mock/mock_repository.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: tidy
tidy:
	$(GOCMD) mod tidy

# Full development cycle
.PHONY: dev-setup
dev-setup: pull-images docker-up migrate

.PHONY: dev-teardown
dev-teardown: docker-down-v clean

# API Documentation
.PHONY: swagger
swagger:
	chmod +x scripts/gen-swagger.sh
	./scripts/gen-swagger.sh

.PHONY: serve-swagger
serve-swagger: swagger
	$(GORUN) ./cmd/api

.PHONY: serve-docs
serve-docs:
	godoc -http=:6060

# Development tools
.PHONY: dev-tools
dev-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/godoc@latest

# Performance profiling
.PHONY: prof-cpu
prof-cpu:
	go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

.PHONY: prof-mem
prof-mem:
	go tool pprof http://localhost:8080/debug/pprof/heap

.PHONY: prof-goroutine
prof-goroutine:
	go tool pprof http://localhost:8080/debug/pprof/goroutine

.PHONY: prof-trace
prof-trace:
	curl -o trace.out http://localhost:8080/debug/pprof/trace?seconds=30
	go tool trace trace.out

.PHONY: prof-flamegraph
prof-flamegraph:
	go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile?seconds=30
