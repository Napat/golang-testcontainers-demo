package handler

import (
	"net/http"

	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_cache"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_event"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_order"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_product"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_user"
)

// Handler holds all HTTP handlers
type Handler struct {
	userHandler    *UserHandler
	productHandler *ProductHandler
	orderHandler   *OrderHandler
	messageHandler *MessageHandler
}

// New creates a new Handler
func New(
	userRepo *repository_user.UserRepository,
	productRepo *repository_product.ProductRepository,
	orderRepo *repository_order.OrderRepository,
	producer *repository_event.ProducerRepository,
	cache *repository_cache.CacheRepository,
) *Handler {
	return &Handler{
		userHandler:    NewUserHandler(userRepo, cache, producer),
		productHandler: NewProductHandler(productRepo),
		orderHandler:   NewOrderHandler(orderRepo),
		messageHandler: NewMessageHandler(producer),
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/users" || r.URL.Path == "/users/":
		h.userHandler.ServeHTTP(w, r)
	case r.URL.Path == "/products" || r.URL.Path == "/products/":
		h.productHandler.ServeHTTP(w, r)
	case r.URL.Path == "/orders" || r.URL.Path == "/orders/":
		h.orderHandler.ServeHTTP(w, r)
	case r.URL.Path == "/messages" || r.URL.Path == "/messages/":
		h.messageHandler.ServeHTTP(w, r)
	default:
		http.NotFound(w, r)
	}
}
