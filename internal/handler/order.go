package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_order"
	"github.com/Napat/golang-testcontainers-demo/pkg/errors"
	"github.com/Napat/golang-testcontainers-demo/pkg/middleware"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
)

type OrderHandler struct {
	repo *repository_order.OrderRepository
}

func NewOrderHandler(repo *repository_order.OrderRepository) *OrderHandler {
	return &OrderHandler{repo: repo}
}

func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/orders":
		h.CreateOrder(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/orders/search":
		h.searchOrders(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/orders/simple-search":
		h.handleOrder(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/orders":
		h.ListOrders(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary Create a new order
// @Description Create a new order with the given details
// @Tags orders
// @Accept json
// @Produce json
// @Param order body model.Order true "Order details"
// @Success 201 {object} model.Order
// @Failure 400 {object} errors.Error
// @Failure 500 {object} errors.Error
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
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

// @Summary Search orders
// @Description Search orders using Elasticsearch query
// @Tags orders
// @Accept json
// @Produce json
// @Param query body object true "Search query"
// @Success 200 {array} model.Order
// @Success 204 "No Content"
// @Failure 400 {object} errors.Error
// @Failure 500 {object} errors.Error
// @Router /orders/search [get]
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

// @Summary Simple search orders
// @Description Search orders using a simple query parameter
// @Tags orders
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} model.Order
// @Success 204 "No Content"
// @Failure 400 {object} errors.Error
// @Failure 500 {object} errors.Error
// @Router /orders/simple-search [get]
func (h *OrderHandler) handleOrder(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		middleware.WriteError(w, errors.NewBadRequest("simpleSearch", "query parameter 'q' is required"))
		return
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"customer_id", "status", "items.product_id", "items.product_name"},
				"type":   "phrase_prefix",
			},
		},
	}

	orders, err := h.repo.SearchOrders(r.Context(), searchQuery)
	if err != nil {
		middleware.WriteError(w, errors.NewInternalError("simpleSearch", err))
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		middleware.WriteError(w, errors.NewInternalError("simpleSearch", err))
	}
}

// @Summary List all orders
// @Description Get a list of all orders
// @Tags orders
// @Accept json
// @Produce json
// @Success 200 {array} model.Order
// @Success 204 "No Content"
// @Failure 500 {object} errors.Error
// @Router /orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.repo.SearchOrders(r.Context(), map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	})
	if err != nil {
		middleware.WriteError(w, errors.NewInternalError("listOrders", err))
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	json.NewEncoder(w).Encode(orders)
}
