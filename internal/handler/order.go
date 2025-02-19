package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Napat/golang-testcontainers-demo/pkg/errors"
	"github.com/Napat/golang-testcontainers-demo/pkg/middleware"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/Napat/golang-testcontainers-demo/pkg/response"
	"github.com/Napat/golang-testcontainers-demo/pkg/routes"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	SearchOrders(ctx context.Context, params map[string]interface{}) ([]model.Order, error)
}

type OrderHandler struct {
	orderRepo OrderRepository
	routes    []routes.Route
}

func NewOrderHandler(repo OrderRepository) *OrderHandler {
	h := &OrderHandler{
		orderRepo: repo,
	}

	h.routes = []routes.Route{
		{
			Method:  http.MethodGet,
			Pattern: "/orders",
			Handler: h.ListOrders,
		},
		{
			Method:  http.MethodPost,
			Pattern: "/orders",
			Handler: h.createOrder,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/orders/search",
			Handler: h.searchOrders,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/orders/simple-search",
			Handler: h.simpleSearch,
		},
	}

	return h
}

// GetRoutes returns all routes for this handler
func (h *OrderHandler) GetRoutes() []routes.Route {
	return h.routes
}

// @Summary Create a new order
// @Description Create a new order in the system
// @Tags orders
// @Accept json
// @Produce json
// @Param order body model.Order true "Order object"
// @Success 201 {object} model.Order
// @Failure 400 {object} map[string]string
// @Router /api/v1/orders [post]
func (h *OrderHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", "createOrder")
		return
	}

	if err := h.orderRepo.CreateOrder(r.Context(), &order); err != nil {
		log.Printf("Error creating order: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "Failed to create order", "createOrder")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// @Summary Search orders
// @Description Search orders using customer ID
// @Tags orders
// @Accept json
// @Produce json
// @Param customer_id query string false "Customer ID to search for"
// @Success 200 {array} model.Order
// @Router /api/v1/orders/search [get]
func (h *OrderHandler) searchOrders(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	if customerID := r.URL.Query().Get("customer_id"); customerID != "" {
		params["customer_id"] = customerID
	}

	orders, err := h.orderRepo.SearchOrders(r.Context(), params)
	if err != nil {
		log.Printf("Error searching orders: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "Failed to search orders", "searchOrders")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// @Summary Simple search orders
// @Description Search orders using a simple query string
// @Tags orders
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} model.Order
// @Router /api/v1/orders/simple-search [get]
func (h *OrderHandler) simpleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		response.RespondWithError(w, http.StatusBadRequest, "Search query is required", "simpleSearch")
		return
	}

	params := map[string]interface{}{
		"query": query,
	}

	orders, err := h.orderRepo.SearchOrders(r.Context(), params)
	if err != nil {
		log.Printf("Error performing simple search: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "Failed to search orders", "simpleSearch")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
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
	orders, err := h.orderRepo.SearchOrders(r.Context(), map[string]interface{}{
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

// ServeHTTP implements http.Handler interface
func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
