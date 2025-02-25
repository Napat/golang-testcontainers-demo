package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
)

type HealthHandler struct {
	mysqlDB     *sql.DB
	postgresDB  *sql.DB
	redisClient *redis.Client
	kafka       sarama.Client
	elastic     *elasticsearch.Client
}

type HealthStatus struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Services  map[string]ServiceStatus `json:"services"`
}

type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

func NewHealthHandler(mysqlDB, postgresDB *sql.DB, redisClient *redis.Client, kafka sarama.Client, elastic *elasticsearch.Client) *HealthHandler {
	return &HealthHandler{
		mysqlDB:     mysqlDB,
		postgresDB:  postgresDB,
		redisClient: redisClient,
		kafka:       kafka,
		elastic:     elastic,
	}
}

// RegisterRoutes registers all health check endpoints
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/health/live", h.handleLive)
	mux.HandleFunc("/health/ready", h.handleReady)
}

func (h *HealthHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.Health(w, r)
}

func (h *HealthHandler) handleLive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Clean up path to handle both with and without trailing slash
	if r.URL.Path != "/health/live" && r.URL.Path != "/health/live/" {
		http.NotFound(w, r)
		return
	}
	h.Live(w, r)
}

func (h *HealthHandler) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Clean up path to handle both with and without trailing slash
	if r.URL.Path != "/health/ready" && r.URL.Path != "/health/ready/" {
		http.NotFound(w, r)
		return
	}
	h.Ready(w, r)
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// @Summary Get health status
// @Description Get the health status of all system components
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthStatus
// @Router /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceStatus),
	}

	// Check MySQL
	status.Services["mysql"] = h.checkMySQL(ctx)

	// Check PostgreSQL
	status.Services["postgres"] = h.checkPostgres(ctx)

	// Check Redis
	status.Services["redis"] = h.checkRedis(ctx)

	// Check Kafka
	status.Services["kafka"] = h.checkKafka()

	// Check Elasticsearch
	status.Services["elasticsearch"] = h.checkElasticsearch(ctx)

	// If any service is down, set overall status to degraded
	for _, s := range status.Services {
		if s.Status != "up" {
			status.Status = "degraded"
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *HealthHandler) checkMySQL(ctx context.Context) ServiceStatus {
	if h.mysqlDB == nil {
		return ServiceStatus{
			Status:  "down",
			Message: "mysql connection not initialized",
		}
	}

	start := time.Now()
	err := h.mysqlDB.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		return ServiceStatus{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return ServiceStatus{
		Status:  "up",
		Latency: latency.String(),
	}
}

func (h *HealthHandler) checkPostgres(ctx context.Context) ServiceStatus {
	if h.postgresDB == nil {
		return ServiceStatus{
			Status:  "down",
			Message: "postgres connection not initialized",
		}
	}

	start := time.Now()
	err := h.postgresDB.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		return ServiceStatus{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return ServiceStatus{
		Status:  "up",
		Latency: latency.String(),
	}
}

func (h *HealthHandler) checkRedis(ctx context.Context) ServiceStatus {
	if h.redisClient == nil {
		return ServiceStatus{
			Status:  "down",
			Message: "redis connection not initialized",
		}
	}

	start := time.Now()
	_, err := h.redisClient.Ping(ctx).Result()
	latency := time.Since(start)

	if err != nil {
		return ServiceStatus{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return ServiceStatus{
		Status:  "up",
		Latency: latency.String(),
	}
}

func (h *HealthHandler) checkKafka() ServiceStatus {
	if h.kafka == nil {
		return ServiceStatus{
			Status:  "down",
			Message: "kafka client not initialized",
		}
	}

	brokers := h.kafka.Brokers()
	if len(brokers) == 0 {
		return ServiceStatus{
			Status:  "down",
			Message: "no kafka brokers available",
		}
	}

	return ServiceStatus{
		Status: "up",
	}
}

func (h *HealthHandler) checkElasticsearch(ctx context.Context) ServiceStatus {
	if h.elastic == nil {
		return ServiceStatus{
			Status:  "down",
			Message: "elasticsearch client not initialized",
		}
	}

	start := time.Now()
	res, err := h.elastic.Ping(
		h.elastic.Ping.WithContext(ctx),
	)
	latency := time.Since(start)

	if err != nil {
		return ServiceStatus{
			Status:  "down",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}
	defer res.Body.Close()

	return ServiceStatus{
		Status:  "up",
		Latency: latency.String(),
	}
}
