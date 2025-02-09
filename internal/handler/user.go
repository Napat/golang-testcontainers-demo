package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Napat/golang-testcontainers-demo/internal/model"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/cache"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/event"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/user"
	"github.com/gorilla/mux"
)

type UserHandler struct {
    userRepo  *user.UserRepository
    cache     *cache.CacheRepository
    producer  *event.ProducerRepository
}

func NewUserHandler(userRepo *user.UserRepository, cache *cache.CacheRepository, producer *event.ProducerRepository) *UserHandler {
    return &UserHandler{
        userRepo:  userRepo,
        cache:     cache,
        producer:  producer,
    }
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    switch {
    case r.Method == http.MethodPost && r.URL.Path == "/users":
        h.createUser(w, r)
    case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/users/"):
        h.getUser(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
    var user model.User
    ctx := r.Context()

    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := h.userRepo.Create(ctx ,&user); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Cache the user
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    if err := h.cache.Set(ctx, cacheKey, user, time.Hour); err != nil {
        log.Printf("Failed to cache user: %v", err)
    }

    // Send to Kafka
    if err := h.producer.SendMessage(cacheKey, user); err != nil {
        log.Printf("Failed to send user to Kafka: %v", err)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    id, err := strconv.ParseInt(vars["id"], 10, 64)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    // Try to get from cache first
    cacheKey := fmt.Sprintf("user:%d", id)
    var user *model.User
    err = h.cache.Get(ctx, cacheKey, &user)
    if err == nil {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(user)
        return
    }

    // If not in cache, get from database
    user, err = h.userRepo.GetByID(ctx, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    // Cache the user for future requests
    if err := h.cache.Set(ctx, cacheKey, user, time.Hour); err != nil {
        log.Printf("Failed to cache user: %v", err)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
