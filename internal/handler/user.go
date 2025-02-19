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
	"github.com/Napat/golang-testcontainers-demo/pkg/response"
	"github.com/Napat/golang-testcontainers-demo/pkg/routes"
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
	routes   []routes.Route
}

func NewUserHandler(userRepo UserRepository, cache CacheRepository, producer MessageProducer) *UserHandler {
	h := &UserHandler{
		userRepo: userRepo,
		cache:    cache,
		producer: producer,
	}

	h.routes = []routes.Route{
		{
			Method:  http.MethodGet,
			Pattern: "/users",
			Handler: h.getAllUsers,
		},
		{
			Method:  http.MethodPost,
			Pattern: "/users",
			Handler: h.createUser,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/users/",
			Handler: h.getUserByID,
		},
	}

	return h
}

// GetRoutes implements routes.Handler interface
func (h *UserHandler) GetRoutes() []routes.Route {
	return h.routes
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
// @Router /api/v1/users [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user model.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", "createUser")
		return
	}

	user.ID = uuid.Must(uuid.NewV7())
	log.Printf("Generated UUIDv7 for new user: %s", user.ID)

	if err := h.userRepo.Create(ctx, &user); err != nil {
		log.Printf("Error creating user: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user: %v", err), "createUser")
		return
	}

	cacheKey := fmt.Sprintf("user:%s", user.ID)
	if err := h.cache.Set(r.Context(), cacheKey, user, time.Hour); err != nil {
		log.Printf("Failed to cache user: %v", err)
	}

	if err := h.producer.SendMessage(cacheKey, user); err != nil {
		log.Printf("Failed to send user to Kafka: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// @Summary Get user by ID
// @Description Get user details by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.User
// @Failure 404 {object} map[string]string "Error response"
// @Failure 500 {object} map[string]string "Error response"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) getUserByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /users/{id}
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid path", "getUserByID")
		return
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format", "getUserByID")
		return
	}

	ctx := r.Context()
	cacheKey := fmt.Sprintf("user:%s", id)

	var user *model.User
	err = h.cache.Get(ctx, cacheKey, &user)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}

	user, err = h.userRepo.GetByID(ctx, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == repository_user.ErrUserNotFound {
			status = http.StatusNotFound
		}
		response.RespondWithError(w, status, err.Error(), "getUserByID")
		return
	}

	if err := h.cache.Set(ctx, cacheKey, user, time.Hour); err != nil {
		log.Printf("Failed to cache user: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Get all users
// @Description Get a list of all users in the system
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} model.User
// @Failure 500 {object} map[string]string "Error response"
// @Router /api/v1/users [get]
func (h *UserHandler) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAll(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error(), "getAllUsers")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ServeHTTP implements http.Handler interface
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Remove /api/v1 prefix if present for both test and production compatibility
	path := strings.TrimPrefix(r.URL.Path, "/api/v1")

	// Find matching route
	for _, route := range h.routes {
		if strings.TrimSuffix(path, "/") == strings.TrimSuffix(route.Pattern, "/") && r.Method == route.Method {
			route.Handler(w, r)
			return
		}
	}

	http.NotFound(w, r)
}
