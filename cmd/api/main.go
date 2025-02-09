package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/IBM/sarama"
	"github.com/Napat/golang-testcontainers-demo/internal/config"
	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/cache"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/event"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/order"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/product"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/user"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func main() {
    // Load configuration from yaml
    configPath := os.Getenv("CONFIG_PATH")
    if configPath == "" {
        configPath = "configs/dev.yaml"
    }

    cfg, err := config.Load(configPath)
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

    // Initialize Kafka producer
    kafkaConfig := sarama.NewConfig()
    kafkaConfig.Producer.Return.Successes = true
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

    // Initialize router
    router := http.NewServeMux()

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

    // Register routes
    router.HandleFunc("/users", userHandler.ServeHTTP)
    router.HandleFunc("/products", productHandler.ServeHTTP)
    router.HandleFunc("/orders", orderHandler.ServeHTTP)
    router.HandleFunc("/orders/search", orderHandler.ServeHTTP)

    // Start server
    port := getEnvOrDefault("PORT", "8080")
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, router))
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
