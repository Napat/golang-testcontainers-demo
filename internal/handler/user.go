package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	repository_cache "github.com/Napat/golang-testcontainers-demo/internal/repository/repository_cache"
	repository_event "github.com/Napat/golang-testcontainers-demo/internal/repository/repository_event"
	repository_user "github.com/Napat/golang-testcontainers-demo/internal/repository/repository_user"
	"github.com/Napat/golang-testcontainers-demo/pkg/errors"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
)

type UserHandler struct {
	userRepo *repository_user.UserRepository
	cache    *repository_cache.CacheRepository
	producer *repository_event.ProducerRepository
}

func NewUserHandler(userRepo *repository_user.UserRepository,
	cache *repository_cache.CacheRepository,
	producer *repository_event.ProducerRepository) *UserHandler {
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
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
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
// @Failure 400 {object} errors.Error
// @Failure 500 {object} errors.Error
// @Router /users [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var req model.UserCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(errors.NewBadRequest("createUser", "Invalid request body"))
		return
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
		Status:   model.StatusActive,
		Version:  1,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
// @Failure 404 {object} errors.Error
// @Failure 500 {object} errors.Error
// @Router /users/{id} [get]
func (h *UserHandler) getUserByID(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := r.Context()

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:%d", id)
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
// @Failure 500 {object} errors.Error
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
