package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/Napat/golang-testcontainers-demo/api/docs"
	"github.com/Napat/golang-testcontainers-demo/internal/config"
	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/internal/handler/health"
	"github.com/Napat/golang-testcontainers-demo/internal/middleware"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/cache"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/event"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/order"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/product"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/user"
	"github.com/Napat/golang-testcontainers-demo/internal/shutdown"
	"github.com/Napat/golang-testcontainers-demo/internal/tracing"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// @title Testcontainers Demo API
// @version 1.0
// @description API documentation for Testcontainers Demo project
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
    // Load configuration from yaml
    cfg, err := config.Load(getEnvOrDefault("CONFIG_PATH", "configs/dev.yaml"))
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize MySQL connection
    mysqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database))
    if err != nil {
        log.Fatalf("Failed to connect to MySQL: %v", err)
    }
    defer mysqlDB.Close()

    // Initialize PostgreSQL connection
    postgresDB, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.PostgreSQL.Host, cfg.PostgreSQL.Port, cfg.PostgreSQL.User, cfg.PostgreSQL.Password, cfg.PostgreSQL.Database))
    if err != nil {
        log.Fatalf("Failed to connect to PostgreSQL: %v", err)
    }
    defer postgresDB.Close()

    // Initialize Redis connection
    redisClient := redis.NewClient(&redis.Options{
        Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
    })
    defer redisClient.Close()

    // Initialize Kafka client and producer
    kafkaConfig := sarama.NewConfig()
    kafkaConfig.Producer.Return.Successes = true
    
    // Create Kafka client for health checks
    kafkaClient, err := sarama.NewClient(cfg.Kafka.Brokers, kafkaConfig)
    if err != nil {
        log.Fatalf("Failed to create Kafka client: %v", err)
    }
    defer kafkaClient.Close()
    
    // Create Kafka producer for actual message production
    kafkaProducer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, kafkaConfig)
    if err != nil {
        log.Fatalf("Failed to create Kafka producer: %v", err)
    }
    defer kafkaProducer.Close()

    // Initialize Elasticsearch client
    esClient, err := elasticsearch.NewClient(elasticsearch.Config{
        Addresses: []string{cfg.Elasticsearch.URL},
    })
    if err != nil {
        log.Fatalf("Error creating Elasticsearch client: %v", err)
    }

    // Create shutdown manager
    shutdownManager := shutdown.NewManager(30 * time.Second)

    // Add shutdown handlers
    shutdownManager.AddHandler(func(ctx context.Context) error {
        return mysqlDB.Close()
    })
    shutdownManager.AddHandler(func(ctx context.Context) error {
        return postgresDB.Close()
    })
    shutdownManager.AddHandler(func(ctx context.Context) error {
        return redisClient.Close()
    })
    shutdownManager.AddHandler(func(ctx context.Context) error {
        return kafkaClient.Close()
    })
    shutdownManager.AddHandler(func(ctx context.Context) error {
        return kafkaProducer.Close()
    })

    // Initialize tracer
    cleanup, err := tracing.InitTracer(
        "testcontainers-demo",
        "http://localhost:14268/api/traces",
    )
    if err != nil {
        log.Fatalf("Failed to initialize tracer: %v", err)
    }
    defer cleanup()

    // Add tracing shutdown to manager
    shutdownManager.AddHandler(func(ctx context.Context) error {
        cleanup()
        return nil
    })

    // Initialize router with middleware chain
    router := setupRouter(mysqlDB, postgresDB, redisClient, kafkaClient, kafkaProducer, esClient, cfg)
    chain := middleware.Profiling()(
        middleware.Swagger()(
            middleware.Tracing("testcontainers-demo")(router),
        ),
    )

    // Create server
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
        Handler:      chain,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Add server shutdown handler
    shutdownManager.AddHandler(func(ctx context.Context) error {
        return srv.Shutdown(ctx)
    })

    // Start server
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    // Perform graceful shutdown
    ctx := context.Background()
    if err := shutdownManager.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
        os.Exit(1)
    }

    log.Println("Server exited properly")
}

func setupRouter(mysqlDB, postgresDB *sql.DB, redisClient *redis.Client, 
    kafkaClient sarama.Client, kafkaProducer sarama.SyncProducer, 
    esClient *elasticsearch.Client, cfg *config.Config) http.Handler {
    mux := http.NewServeMux()

    // Initialize repositories
    userRepo := user.NewUserRepository(mysqlDB)
    productRepo := product.NewProductRepository(postgresDB)
    cacheRepo := cache.NewCacheRepository(redisClient)
    kafkaRepo := event.NewProducerRepository(kafkaProducer, cfg.Kafka.Topic)
    orderRepo := order.NewOrderRepository(esClient)

    // Initialize handlers
    userHandler := handler.NewUserHandler(userRepo, cacheRepo, kafkaRepo)
    productHandler := handler.NewProductHandler(productRepo)
    orderHandler := handler.NewOrderHandler(orderRepo)

    // Add health check endpoints with kafka client
    healthHandler := health.NewHealthHandler(
        mysqlDB, 
        postgresDB, 
        redisClient, 
        kafkaClient,
        esClient,
    )
    mux.Handle("/health", healthHandler)
    mux.Handle("/health/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }))
    mux.Handle("/health/ready", healthHandler)

    // Register routes
    mux.HandleFunc("/users", userHandler.ServeHTTP)
    mux.HandleFunc("/products", productHandler.ServeHTTP)
    mux.HandleFunc("/orders", orderHandler.ServeHTTP)
    mux.HandleFunc("/orders/search", orderHandler.ServeHTTP)
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    return mux
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
