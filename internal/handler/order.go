package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Napat/golang-testcontainers-demo/internal/errors"
	"github.com/Napat/golang-testcontainers-demo/internal/middleware"
	"github.com/Napat/golang-testcontainers-demo/internal/model"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/order"
)

type OrderHandler struct {
    repo *order.OrderRepository
}

func NewOrderHandler(repo *order.OrderRepository) *OrderHandler {
    return &OrderHandler{repo: repo}
}

func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    switch {
    case r.Method == http.MethodPost && r.URL.Path == "/orders":
        h.createOrder(w, r)
    case r.Method == http.MethodGet && r.URL.Path == "/orders/search":
        h.searchOrders(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request) {
    var order model.Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        middleware.WriteError(w, errors.NewBadRequest("createOrder", "invalid request body"))
        return
    }

    if err := order.Validate(); err != nil {
        middleware.WriteError(w, errors.NewValidationError("createOrder", err.Error()))
        return
    }

    if err := h.repo.CreateOrder(r.Context(), &order); err != nil {
        middleware.WriteError(w, errors.NewInternalError("createOrder", err))
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) searchOrders(w http.ResponseWriter, r *http.Request) {
    var query map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
        middleware.WriteError(w, errors.NewBadRequest("searchOrders", "invalid query format"))
        return
    }

    orders, err := h.repo.SearchOrders(r.Context(), query)
    if err != nil {
        middleware.WriteError(w, errors.NewInternalError("searchOrders", err))
        return
    }

    if len(orders) == 0 {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    json.NewEncoder(w).Encode(orders)
}
