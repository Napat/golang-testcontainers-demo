package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/Napat/golang-testcontainers-demo/pkg/response"
	"github.com/Napat/golang-testcontainers-demo/pkg/routes"
)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetAll(ctx context.Context) ([]*model.Product, error)
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id int64) error
}

type ProductHandler struct {
	productRepo ProductRepository
	routes      []routes.Route
}

func NewProductHandler(repo ProductRepository) *ProductHandler {
	h := &ProductHandler{
		productRepo: repo,
	}

	h.routes = []routes.Route{
		{
			Method:  http.MethodGet,
			Pattern: "/products",
			Handler: h.getAllProducts,
		},
		{
			Method:  http.MethodPost,
			Pattern: "/products",
			Handler: h.createProduct,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/products/",
			Handler: h.getProductByID,
		},
	}

	return h
}

// GetRoutes returns all routes for this handler
func (h *ProductHandler) GetRoutes() []routes.Route {
	return h.routes
}

// @Summary Get all products
// @Description Get a list of all products
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {array} model.Product
// @Router /api/v1/products [get]
func (h *ProductHandler) getAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productRepo.GetAll(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error(), "getAllProducts")
		return
	}
	response.RespondWithJSON(w, http.StatusOK, products)
}

// @Summary Create a new product
// @Description Create a new product in the system
// @Tags products
// @Accept json
// @Produce json
// @Param product body model.Product true "Product object"
// @Success 201 {object} model.Product
// @Router /api/v1/products [post]
func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var product model.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", "createProduct")
		return
	}

	if err := h.productRepo.Create(r.Context(), &product); err != nil {
		log.Printf("Error creating product: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "Failed to create product", "createProduct")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// @Summary Get product by ID
// @Description Get a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} model.Product
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) getProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/v1/products/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Invalid product ID", "getProductByID")
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), id)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, err.Error(), "getProductByID")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, product)
}

// ServeHTTP implements http.Handler interface
func (h *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
