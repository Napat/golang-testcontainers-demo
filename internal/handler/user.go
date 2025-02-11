package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	repository_user "github.com/Napat/golang-testcontainers-demo/internal/repository/repository_user"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetAll(ctx context.Context) ([]*model.User, error)
}

type CacheRepository interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

type UserHandler struct {
	userRepo UserRepository
	cache    CacheRepository
	producer MessageProducer
}

func respondWithError(w http.ResponseWriter, code int, message string, source string) {
	response := map[string]string{
		"error":  message,
		"source": source,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func NewUserHandler(userRepo UserRepository,
	cache CacheRepository,
	producer MessageProducer) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		cache:    cache,
		producer: producer,
	}
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Trim trailing slash and split path
	path := strings.TrimSuffix(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch {
	case r.Method == http.MethodPost && path == "/users":
		h.createUser(w, r)
	case r.Method == http.MethodGet && path == "/users":
		h.getAllUsers(w, r)
	case r.Method == http.MethodGet && len(parts) == 3 && parts[1] == "users":
		id, err := uuid.Parse(parts[2])
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}
		h.getUserByID(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.UserCreate true "User creation request"
// @Success 201 {object} model.User
// @Failure 400 {object} map[string]string "Error response"
// @Failure 500 {object} map[string]string "Error response"
// @Router /users [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user model.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", "createUser")
		return
	}

	// Generate and log new UUIDv7
	user.ID = uuid.Must(uuid.NewV7())
	log.Printf("Generated UUIDv7 for new user: %s", user.ID)

	if err := h.userRepo.Create(ctx, &user); err != nil {
		log.Printf("Error creating user: %v", err)
		// ระบุ error message ที่ชัดเจนขึ้น
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %v", err), "createUser")
		return
	}

	// Cache the user
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	if err := h.cache.Set(r.Context(), cacheKey, user, time.Hour); err != nil {
		log.Printf("Failed to cache user: %v", err)
	}

	// Send to Kafka
	if err := h.producer.SendMessage(cacheKey, user); err != nil {
		log.Printf("Failed to send user to Kafka: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Get user by ID
// @Description Get user details by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} model.User
// @Failure 404 {object} map[string]string "Error response"
// @Failure 500 {object} map[string]string "Error response"
// @Router /users/{id} [get]
func (h *UserHandler) getUserByID(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	ctx := r.Context()

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:%s", id)
	var user *model.User
	err := h.cache.Get(ctx, cacheKey, &user)
	if err == nil {
		json.NewEncoder(w).Encode(user)
		return
	}

	// If not in cache, get from database
	user, err = h.userRepo.GetByID(ctx, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == repository_user.ErrUserNotFound { // Fix error reference
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	// Cache the user for future requests
	if err := h.cache.Set(ctx, cacheKey, user, time.Hour); err != nil {
		log.Printf("Failed to cache user: %v", err)
	}

	json.NewEncoder(w).Encode(user)
}

// @Summary Get all users
// @Description Get a list of all users in the system
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} model.User
// @Failure 500 {object} map[string]string "Error response"
// @Router /users [get]
func (h *UserHandler) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
