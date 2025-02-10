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
	"github.com/Napat/golang-testcontainers-demo/internal/router"
	"github.com/Napat/golang-testcontainers-demo/pkg/middleware"
	"github.com/Napat/golang-testcontainers-demo/pkg/shutdown"
	"github.com/Napat/golang-testcontainers-demo/pkg/tracing"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Testcontainers Demo API
// @version         1.0
// @description     This is a sample server using testcontainers.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

func main() {
	// Load configuration from yaml
	cfg, err := config.Load(getEnvOrDefault("CONFIG_PATH", "configs/dev.yaml"))
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MySQL connection with parseTime=true
	mysqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
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
	if cfg.Tracing.Enabled {
		cleanup, err := tracing.InitTracer(
			cfg.Tracing.ServiceName,
			cfg.Tracing.CollectorURL,
			tracing.WithSamplingRatio(cfg.Tracing.SamplingRatio),
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
	}

	// Create new mux to register Swagger endpoint And wrap of app endpoints
	rootMux := http.NewServeMux()

	// Add swagger endpoints
	rootMux.Handle("/swagger.json", http.FileServer(http.Dir("api/docs")))
	rootMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Add app endpoints
	rtr := router.Setup(mysqlDB, postgresDB, redisClient, kafkaClient, kafkaProducer, esClient, cfg)
	rootMux.Handle("/", rtr)

	// Add only Profiling and Tracing middleware
	chain := middleware.Profiling()(
		middleware.Tracing("testcontainers-demo")(rootMux),
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

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
