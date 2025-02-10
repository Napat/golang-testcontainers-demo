package router

import (
	"database/sql"
	"net/http"

	"github.com/IBM/sarama"
	"github.com/Napat/golang-testcontainers-demo/internal/config"
	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/internal/handler/health"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_cache"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_event"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_order"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_product"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_user"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
)

func Setup(mysqlDB, postgresDB *sql.DB, redisClient *redis.Client,
	kafkaClient sarama.Client, kafkaProducer sarama.SyncProducer,
	esClient *elasticsearch.Client, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Initialize repositories
	userRepo := repository_user.NewUserRepository(mysqlDB)
	productRepo := repository_product.NewProductRepository(postgresDB)
	cacheRepo := repository_cache.NewCacheRepository(redisClient)
	kafkaRepo := repository_event.NewProducerRepository(kafkaProducer, cfg.Kafka.Topic)
	orderRepo := repository_order.NewOrderRepository(esClient)

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
	health.RegisterHealthRoutes(mux, healthHandler)

	// Register routes
	mux.HandleFunc("/users/", userHandler.ServeHTTP)
	mux.HandleFunc("/users", userHandler.ServeHTTP)
	mux.HandleFunc("/products", productHandler.ServeHTTP)
	mux.HandleFunc("/orders", orderHandler.ServeHTTP)
	mux.HandleFunc("/orders/search", orderHandler.ServeHTTP)
	mux.HandleFunc("/orders/simple-search", orderHandler.ServeHTTP)
	mux.HandleFunc("/search", orderHandler.ServeHTTP)

	// Add message handler
	messageHandler := handler.NewMessageHandler(kafkaRepo)
	mux.HandleFunc("/messages", messageHandler.ServeHTTP)

	return mux
}
