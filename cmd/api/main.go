package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/Napat/golang-testcontainers-demo/api/docs"
	"github.com/Napat/golang-testcontainers-demo/internal/config"
	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/internal/handler/health"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_cache"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_event"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_order"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_product"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_user"
	"github.com/Napat/golang-testcontainers-demo/internal/router"
	"github.com/Napat/golang-testcontainers-demo/pkg/shutdown"
	"github.com/Napat/golang-testcontainers-demo/pkg/tracing"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	buildTime    string // จะถูกกำหนดค่าตอน build ผ่าน ldflags
	gitCommitSHA string // จะถูกกำหนดค่าตอน build ผ่าน ldflags
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

// printEnvironmentInfo แสดงข้อมูล environment และ configuration
func printEnvironmentInfo(configPath string, cfg *config.Config) {
	log.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📦 Testcontainers Demo API")
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	env := getEnvOrDefault("GO_ENV", "development")
	isProd := env == "production"

	log.Printf("📚 Environment:")
	log.Printf("   ├── Mode: %s %s", env, getEnvironmentIcon(isProd))
	log.Printf("   ├── Config: %s", configPath)
	log.Printf("   └── Version: %s\n", "1.0.0")
}

// printSystemInfo แสดงข้อมูลระบบ
func printSystemInfo() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("\n💻 System Information:")
	log.Printf("   ├── Go Version: %s", runtime.Version())
	log.Printf("   ├── GOOS/GOARCH: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("   ├── CPU Cores: %d", runtime.NumCPU())
	log.Printf("   ├── Goroutines: %d", runtime.NumGoroutine())
	log.Printf("   └── Memory:")
	log.Printf("       ├── Alloc: %.2f MB", float64(m.Alloc)/1024/1024)
	log.Printf("       ├── TotalAlloc: %.2f MB", float64(m.TotalAlloc)/1024/1024)
	log.Printf("       └── Sys: %.2f MB", float64(m.Sys)/1024/1024)
}

// printServerInfo แสดงข้อมูลเซิร์ฟเวอร์
func printServerInfo(cfg *config.Config, addr string, startTime time.Time) {
	log.Printf("\n🚀 Server Details:")
	log.Printf("   ├── Environment: %s", cfg.Environment)
	log.Printf("   ├── Port: %s", cfg.Server.Port)
	log.Printf("   ├── Startup Time: %v", time.Since(startTime).Round(time.Millisecond))
	log.Printf("   └── PID: %d", os.Getpid())

	log.Printf("\n🌐 Access URLs:")
	log.Printf("   ├── Local API: http://localhost%s", addr)
	log.Printf("   ├── Health Check: http://localhost%s/health", addr)
	log.Printf("   ├── Metrics: http://localhost%s/metrics", addr)
	log.Printf("   ├── Swagger UI: http://localhost%s/swagger/", addr)
	log.Printf("   └── Swagger JSON: http://localhost%s/swagger.json", addr)
}

// printEndpoints แสดงรายการ endpoints ที่มี
func printEndpoints() {
	log.Printf("\n📝 API Endpoints:")
	log.Printf("   ├── Health:")
	log.Printf("   │   └── GET    /health                - Health check")
	log.Printf("   ├── Users:")
	log.Printf("   │   ├── GET    /api/v1/users         - List users")
	log.Printf("   │   └── POST   /api/v1/users         - Create user")
	log.Printf("   ├── Products:")
	log.Printf("   │   └── GET    /api/v1/products      - List products")
	log.Printf("   ├── Orders:")
	log.Printf("   │   ├── POST   /api/v1/orders        - Create order")
	log.Printf("   │   └── GET    /api/v1/orders/search - Search orders")
	log.Printf("   └── Messages:")
	log.Printf("       └── POST   /api/v1/messages      - Send message")
}

// printDevTools แสดงรายการเครื่องมือสำหรับ development
func printDevTools(cfg *config.Config, addr string) {
	if cfg.Tracing.Enabled {
		log.Printf("\n🔍 Debug Tools:")
		log.Printf("   ├── pprof: http://localhost%s/debug/pprof", addr)
		log.Printf("   ├── Heap: http://localhost%s/debug/pprof/heap", addr)
		log.Printf("   └── Goroutines: http://localhost%s/debug/pprof/goroutine", addr)
	}

	log.Printf("\n💡 Helper Commands:")
	log.Printf("   ├── make test          - Run tests")
	log.Printf("   ├── make coverage      - Generate test coverage")
	log.Printf("   ├── make migrate       - Run database migrations")
	log.Printf("   └── make docs          - Generate API documentation")
}

// getEnvironmentIcon returns emoji based on environment
func getEnvironmentIcon(isProd bool) string {
	if isProd {
		return "🔒"
	}
	return "🔧"
}

// initializeServices ทำการเชื่อมต่อกับ services ต่างๆ
func initializeServices(cfg *config.Config) (*sql.DB, *sql.DB, *redis.Client, sarama.Client, sarama.SyncProducer, *elasticsearch.Client, error) {
	var criticalError error

	// MySQL - ถ้า error ถือว่าเป็น critical
	mysqlDB, err := initializeMySQLConnection(cfg)
	if err != nil {
		criticalError = fmt.Errorf("mysql init failed: %w", err)
		return nil, nil, nil, nil, nil, nil, criticalError
	}

	// PostgreSQL - ถ้า error ถือว่าเป็น critical
	postgresDB, err := initializePostgreSQLConnection(cfg)
	if err != nil {
		criticalError = fmt.Errorf("postgres init failed: %w", err)
		return mysqlDB, nil, nil, nil, nil, nil, criticalError
	}

	// Redis - ไม่ถือว่าเป็น critical error
	redisClient, err := initializeRedisConnection(cfg)
	if err != nil {
		log.Printf("⚠️  Redis initialization warning: %v", err)
	}

	// Kafka - ไม่ถือว่าเป็น critical error
	kafkaClient, kafkaProducer, err := initializeKafkaConnection(cfg)
	if err != nil {
		log.Printf("⚠️  Kafka initialization warning: %v", err)
	}

	// Elasticsearch - ไม่ถือว่าเป็น critical error
	esClient, err := initializeElasticsearchConnection(cfg)
	if err != nil {
		log.Printf("⚠️  Elasticsearch initialization warning: %v", err)
	}

	return mysqlDB, postgresDB, redisClient, kafkaClient, kafkaProducer, esClient, nil
}

func initializeMySQLConnection(cfg *config.Config) (*sql.DB, error) {
	log.Printf("   ├── MySQL:")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database))
	if err != nil {
		log.Printf("   │   └── ❌ Connection failed: %v", err)
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.MySQL.MaxLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(cfg.MySQL.MaxIdleTime) * time.Minute)

	if err := db.Ping(); err != nil {
		log.Printf("   │   └── ⚠️  Connected but ping failed: %v", err)
		return db, err
	}

	log.Printf("   │   ├── ✅ Connected successfully to %s:%s/%s",
		cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
	log.Printf("   │   └── 🔄 Max lifetime: 5m, Max idle: 5m")
	return db, nil
}

func initializePostgreSQLConnection(cfg *config.Config) (*sql.DB, error) {
	log.Printf("   ├── PostgreSQL:")
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgreSQL.Host, cfg.PostgreSQL.Port, cfg.PostgreSQL.User,
		cfg.PostgreSQL.Password, cfg.PostgreSQL.Database))
	if err != nil {
		log.Printf("   │   └── ❌ Connection failed: %v", err)
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.PostgreSQL.MaxOpenConns)
	db.SetMaxIdleConns(cfg.PostgreSQL.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.PostgreSQL.MaxLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(cfg.PostgreSQL.MaxIdleTime) * time.Minute)

	if err := db.Ping(); err != nil {
		log.Printf("   │   └── ⚠️  Connected but ping failed: %v", err)
		return db, err
	}

	log.Printf("   │   ├── ✅ Connected successfully to %s:%s/%s",
		cfg.PostgreSQL.Host, cfg.PostgreSQL.Port, cfg.PostgreSQL.Database)
	log.Printf("   │   └── 🔄 Max lifetime: 5m, Max idle: 5m")
	return db, nil
}

func initializeRedisConnection(cfg *config.Config) (*redis.Client, error) {
	log.Printf("   ├── Redis:")
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Printf("   │   └── ❌ Connection failed: %v", err)
		return client, err
	}

	log.Printf("   │   └── ✅ Connected successfully to %s:%s",
		cfg.Redis.Host, cfg.Redis.Port)
	return client, nil
}

func initializeKafkaConnection(cfg *config.Config) (sarama.Client, sarama.SyncProducer, error) {
	log.Printf("   ├── Kafka:")
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	client, err := sarama.NewClient(cfg.Kafka.Brokers, config)
	if err != nil {
		log.Printf("   │   ├── ❌ Connection failed: %v", err)
		log.Printf("   │   └── Brokers: %v", cfg.Kafka.Brokers)
		return nil, nil, err
	}

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, config)
	if err != nil {
		log.Printf("   │   ├── ⚠️  Connected but producer creation failed: %v", err)
		log.Printf("   │   ├── Brokers: %v", cfg.Kafka.Brokers)
		log.Printf("   │   └── Topic: %s", cfg.Kafka.Topic)
		return client, nil, err
	}

	log.Printf("   │   ├── ✅ Connected successfully")
	log.Printf("   │   ├── Brokers: %v", cfg.Kafka.Brokers)
	log.Printf("   │   └── Topic: %s", cfg.Kafka.Topic)
	return client, producer, nil
}

func initializeElasticsearchConnection(cfg *config.Config) (*elasticsearch.Client, error) {
	log.Printf("   └── Elasticsearch:")
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.Elasticsearch.URL},
	})
	if err != nil {
		log.Printf("       └── ❌ Client creation failed: %v", err)
		return nil, err
	}

	info, err := client.Info()
	if err != nil {
		log.Printf("       └── ⚠️  Client created but ping failed: %v", err)
		return client, err
	}
	defer info.Body.Close()

	log.Printf("       └── ✅ Connected successfully to %s", cfg.Elasticsearch.URL)
	return client, nil
}

// setupGracefulShutdown จัดการการปิดระบบอย่างสมบูรณ์
func setupGracefulShutdown(
	srv *http.Server,
	mysqlDB *sql.DB,
	postgresDB *sql.DB,
	redisClient *redis.Client,
	kafkaProducer sarama.SyncProducer,
	kafkaClient sarama.Client,
) func() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("\n🛑 Initiating graceful shutdown...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var shutdownErrs []error

		// 1. First shutdown HTTP server
		if err := srv.Shutdown(ctx); err != nil {
			shutdownErrs = append(shutdownErrs, fmt.Errorf("server shutdown: %w", err))
		} else {
			log.Println("✅ HTTP server stopped")
		}

		// 2. Then close Redis (non-critical)
		if redisClient != nil {
			if err := redisClient.Close(); err != nil {
				shutdownErrs = append(shutdownErrs, fmt.Errorf("redis cleanup: %w", err))
			} else {
				log.Println("✅ Redis connection closed")
			}
		}

		// 3. Close Kafka producer first, then client
		if kafkaProducer != nil {
			// Use a sync.Once to ensure we only close once
			var once sync.Once
			once.Do(func() {
				if err := kafkaProducer.Close(); err != nil {
					shutdownErrs = append(shutdownErrs, fmt.Errorf("kafka producer cleanup: %w", err))
				} else {
					log.Println("✅ Kafka producer closed")
				}
			})
			// Wait a bit before closing the client
			time.Sleep(time.Second)
		}

		if kafkaClient != nil {
			if err := kafkaClient.Close(); err != nil {
				shutdownErrs = append(shutdownErrs, fmt.Errorf("kafka client cleanup: %w", err))
			} else {
				log.Println("✅ Kafka client closed")
			}
		}

		// 4. Finally close databases
		if postgresDB != nil {
			if err := postgresDB.Close(); err != nil {
				shutdownErrs = append(shutdownErrs, fmt.Errorf("postgres cleanup: %w", err))
			} else {
				log.Println("✅ PostgreSQL connection closed")
			}
		}

		if mysqlDB != nil {
			if err := mysqlDB.Close(); err != nil {
				shutdownErrs = append(shutdownErrs, fmt.Errorf("mysql cleanup: %w", err))
			} else {
				log.Println("✅ MySQL connection closed")
			}
		}

		// Report any shutdown errors
		if len(shutdownErrs) > 0 {
			log.Println("\n⚠️  Shutdown completed with errors:")
			for _, err := range shutdownErrs {
				log.Printf("   - %v", err)
			}
		} else {
			log.Println("\n✅ All connections closed successfully")
		}

		log.Println("👋 Server shutdown complete")
		os.Exit(0)
	}()

	return func() {
		log.Println("\n🔄 Running cleanup handlers...")
		quit <- syscall.SIGTERM
	}
}

func printConnectionPoolInfo(mysqlDB *sql.DB, postgresDB *sql.DB, redisClient *redis.Client) {
	log.Printf("\n📊 Connection Pools:")

	// MySQL pool stats
	log.Printf("   ├── MySQL:")
	log.Printf("   │   ├── Max Open: %d", mysqlDB.Stats().MaxOpenConnections)
	log.Printf("   │   ├── Open: %d", mysqlDB.Stats().OpenConnections)
	log.Printf("   │   ├── In Use: %d", mysqlDB.Stats().InUse)
	log.Printf("   │   └── Idle: %d", mysqlDB.Stats().Idle)

	// PostgreSQL pool stats
	log.Printf("   ├── PostgreSQL:")
	log.Printf("   │   ├── Max Open: %d", postgresDB.Stats().MaxOpenConnections)
	log.Printf("   │   ├── Open: %d", postgresDB.Stats().OpenConnections)
	log.Printf("   │   ├── In Use: %d", postgresDB.Stats().InUse)
	log.Printf("   │   └── Idle: %d", postgresDB.Stats().Idle)

	// Redis pool stats
	log.Printf("   └── Redis:")
	poolStats := redisClient.PoolStats()
	log.Printf("       ├── Total Conns: %d", poolStats.TotalConns)
	log.Printf("       ├── Idle Conns: %d", poolStats.IdleConns)
	log.Printf("       └── Stale Conns: %d", poolStats.StaleConns)
}

// getAppVersion returns application version from build info
func getAppVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "development"
}

// getDependencyVersions returns important dependencies versions
func getDependencyVersions() map[string]string {
	versions := make(map[string]string)
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			switch {
			case strings.Contains(dep.Path, "go-sql-driver/mysql"):
				versions["mysql"] = dep.Version
			case strings.Contains(dep.Path, "lib/pq"):
				versions["postgres"] = dep.Version
			case strings.Contains(dep.Path, "go-redis"):
				versions["redis"] = dep.Version
			case strings.Contains(dep.Path, "sarama"):
				versions["kafka"] = dep.Version
			case strings.Contains(dep.Path, "elasticsearch"):
				versions["elasticsearch"] = dep.Version
			}
		}
	}
	return versions
}

// printVersionInfo แสดงข้อมูลเวอร์ชันและการ build
func printVersionInfo() {
	versions := getDependencyVersions()

	log.Printf("\n📦 Build Information:")
	log.Printf("   ├── Version: %s", getAppVersion())
	if buildTime != "" {
		log.Printf("   ├── Build Time: %s", buildTime)
	}
	if gitCommitSHA != "" {
		log.Printf("   ├── Git Commit: %s", gitCommitSHA)
	}
	log.Printf("   └── Dependencies:")
	log.Printf("       ├── MySQL Driver: %s", versions["mysql"])
	log.Printf("       ├── PostgreSQL Driver: %s", versions["postgres"])
	log.Printf("       ├── Redis Client: %s", versions["redis"])
	log.Printf("       ├── Kafka Client: %s", versions["kafka"])
	log.Printf("       └── Elasticsearch Client: %s", versions["elasticsearch"])
}

// printActiveConfiguration แสดงการตั้งค่าที่ใช้งานอยู่
func printActiveConfiguration(cfg *config.Config, srv *http.Server) {
	log.Printf("\n⚙️  Active Configuration:")

	// Server timeouts
	log.Printf("   ├── Timeouts:")
	log.Printf("   │   ├── Read: %v", srv.ReadTimeout)
	log.Printf("   │   ├── Write: %v", srv.WriteTimeout)
	log.Printf("   │   └── Idle: %v", srv.IdleTimeout)

	// Connection pools
	log.Printf("   ├── Connection Pools:")
	log.Printf("   │   ├── MySQL Max Conn: %d", cfg.MySQL.MaxOpenConns)
	log.Printf("   │   ├── PostgreSQL Max Conn: %d", cfg.PostgreSQL.MaxOpenConns)
	log.Printf("   │   └── Redis Pool Size: %d", cfg.Redis.PoolSize)

	// Tracing settings if enabled
	if cfg.Tracing.Enabled {
		log.Printf("   ├── Tracing:")
		log.Printf("   │   ├── Service: %s", cfg.Tracing.ServiceName)
		log.Printf("   │   ├── Collector: %s", cfg.Tracing.CollectorURL)
		log.Printf("   │   └── Sampling Ratio: %.2f", cfg.Tracing.SamplingRatio)
	}

	// Environment variables
	log.Printf("   └── Environment Variables:")
	log.Printf("       ├── GO_ENV: %s", getEnvOrDefault("GO_ENV", "development"))
	log.Printf("       ├── CONFIG_PATH: %s", getEnvOrDefault("CONFIG_PATH", "configs/dev.yaml"))
	log.Printf("       └── PORT: %s", cfg.Server.Port)
}

func main() {
	startTime := time.Now()

	// Load configuration
	configPath := getEnvOrDefault("CONFIG_PATH", "configs/dev.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Print initial information
	printEnvironmentInfo(configPath, cfg)
	printVersionInfo()

	log.Printf("\n🔌 Services Status:")

	// Initialize services
	mysqlDB, postgresDB, redisClient, kafkaClient, kafkaProducer, esClient, err := initializeServices(cfg)
	if err != nil {
		log.Fatalf("❌ Critical service initialization failed: %v", err)
	}

	// Initialize repositories and handlers
	userRepo := repository_user.NewUserRepository(mysqlDB)
	productRepo := repository_product.NewProductRepository(postgresDB)
	cacheRepo := repository_cache.NewCacheRepository(redisClient)
	eventRepo := repository_event.NewProducerRepository(kafkaProducer, cfg.Kafka.Topic)
	orderRepo := repository_order.NewOrderRepository(esClient)
	healthHandler := health.NewHealthHandler(mysqlDB, postgresDB, redisClient, kafkaClient, esClient)

	// Setup deferred cleanup
	defer func() {
		if mysqlDB != nil {
			mysqlDB.Close()
		}
		if postgresDB != nil {
			postgresDB.Close()
		}
		if redisClient != nil {
			redisClient.Close()
		}
		if kafkaClient != nil {
			kafkaClient.Close()
		}
	}()

	// Print system information
	printSystemInfo()

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	printServerInfo(cfg, addr, startTime)
	printEndpoints()
	printDevTools(cfg, addr)

	// Print connection pools info
	printConnectionPoolInfo(mysqlDB, postgresDB, redisClient)

	// Initialize shutdown manager
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

	// Initialize tracer if enabled
	if cfg.Tracing.Enabled {
		log.Printf("\n🔍 Tracing:")
		log.Printf("   ├── Service: %s", cfg.Tracing.ServiceName)
		log.Printf("   ├── Collector: %s", cfg.Tracing.CollectorURL)
		log.Printf("   └── Sampling Ratio: %.2f", cfg.Tracing.SamplingRatio)

		cleanup, err := tracing.InitTracer(
			cfg.Tracing.ServiceName,
			cfg.Tracing.CollectorURL,
			tracing.WithSamplingRatio(cfg.Tracing.SamplingRatio),
		)
		if err != nil {
			log.Fatalf("❌ Failed to initialize tracer: %v", err)
		}
		defer cleanup()

		shutdownManager.AddHandler(func(ctx context.Context) error {
			cleanup()
			return nil
		})
	}

	// Initialize handlers
	userHandler := handler.NewUserHandler(userRepo, cacheRepo, eventRepo)
	productHandler := handler.NewProductHandler(productRepo)
	orderHandler := handler.NewOrderHandler(orderRepo)
	messageHandler := handler.NewMessageHandler(eventRepo)

	// Setup router using the router package
	routerHandler := router.Setup(
		userHandler,
		productHandler,
		orderHandler,
		messageHandler,
		healthHandler,
		cfg,
	)

	// Setup HTTP server
	rootMux := http.NewServeMux()

	// Add swagger endpoints
	rootMux.Handle("/swagger.json", http.FileServer(http.Dir("api/docs")))
	rootMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Add metrics endpoint
	rootMux.Handle("/metrics", promhttp.Handler())

	// Mount router
	rootMux.Handle("/", routerHandler)

	// Create and configure server
	srv := &http.Server{
		Addr:         addr,
		Handler:      rootMux,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Setup graceful shutdown
	cleanup := setupGracefulShutdown(srv, mysqlDB, postgresDB, redisClient, kafkaProducer, kafkaClient)
	defer cleanup()

	// Add configuration details
	printActiveConfiguration(cfg, srv)

	log.Printf("\n✨ Server is ready to accept connections")
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Start server
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
