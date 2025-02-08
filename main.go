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
	"github.com/Napat/golang-testcontainers-demo/internal/repository/product"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/user"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
    // Load configuration
    cfg := loadConfig()

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

    // Initialize repositories
    userRepo := user.NewUserRepository(mysqlDB)
    productRepo := product.NewProductRepository(postgresDB)
    cacheRepo := cache.NewCacheRepository(redisClient)
    kafkaRepo := event.NewProducerRepository(kafkaProducer, cfg.Kafka.Topic)

    // Initialize handlers
    userHandler := handler.NewUserHandler(userRepo, cacheRepo, kafkaRepo)
    productHandler := handler.NewProductHandler(productRepo)

    // Set up router
    router := mux.NewRouter()

    // User routes
    router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
    router.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")

    // Product routes
    router.HandleFunc("/products", productHandler.CreateProduct).Methods("POST")
    router.HandleFunc("/products/{id}", productHandler.GetProduct).Methods("GET")

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, router))
}

func loadConfig() *config.Config {
    // In a real application, this would load from environment variables or config file
    cfg := &config.Config{}
    
    // MySQL Config
    cfg.MySQL.Host = getEnvOrDefault("MYSQL_HOST", "localhost")
    cfg.MySQL.Port = getEnvOrDefault("MYSQL_PORT", "3306")
    cfg.MySQL.User = getEnvOrDefault("MYSQL_USER", "root")
    cfg.MySQL.Password = getEnvOrDefault("MYSQL_PASSWORD", "password")
    cfg.MySQL.Database = getEnvOrDefault("MYSQL_DATABASE", "testdb")

    // PostgreSQL Config
    cfg.PostgreSQL.Host = getEnvOrDefault("POSTGRES_HOST", "localhost")
    cfg.PostgreSQL.Port = getEnvOrDefault("POSTGRES_PORT", "5432")
    cfg.PostgreSQL.User = getEnvOrDefault("POSTGRES_USER", "postgres")
    cfg.PostgreSQL.Password = getEnvOrDefault("POSTGRES_PASSWORD", "password")
    cfg.PostgreSQL.Database = getEnvOrDefault("POSTGRES_DATABASE", "testdb")

    // Redis Config
    cfg.Redis.Host = getEnvOrDefault("REDIS_HOST", "localhost")
    cfg.Redis.Port = getEnvOrDefault("REDIS_PORT", "6379")

    // Kafka Config
    cfg.Kafka.Brokers = []string{getEnvOrDefault("KAFKA_BROKER", "localhost:9092")}
    cfg.Kafka.Topic = getEnvOrDefault("KAFKA_TOPIC", "test-topic")

    return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
