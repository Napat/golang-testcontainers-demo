package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Napat/golang-testcontainers-demo/pkg/errors"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	List(ctx context.Context) ([]*model.Product, error)
}

type ProductHandler struct {
	productRepo ProductRepository
}

func NewProductHandler(productRepo ProductRepository) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
	}
}

func (h *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/products":
		h.createProduct(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/products":
		h.listProducts(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/products/"):
		h.getProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// @Summary Create a new product
// @Description Create a new product with the provided details
// @Tags products
// @Accept json
// @Produce json
// @Param product body model.Product true "Product details"
// @Success 201 {object} model.Product
// @Failure 400 {object} errors.Error
// @Failure 500 {object} errors.Error
// @Router /products [post]
func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var product model.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		json.NewEncoder(w).Encode(errors.NewBadRequest("createProduct", "Invalid request body"))
		return
	}

	if err := h.productRepo.Create(r.Context(), &product); err != nil {
		json.NewEncoder(w).Encode(errors.NewInternalError("createProduct", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// @Summary Get product by ID
// @Description Get product details by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} model.Product
// @Failure 404 {object} errors.Error
// @Failure 500 {object} errors.Error
// @Router /products/{id} [get]
func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/products/")

	// Convert string ID to int64
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		json.NewEncoder(w).Encode(errors.NewBadRequest("getProduct", "Invalid product ID"))
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil {
		json.NewEncoder(w).Encode(errors.NewInternalError("getProduct", err))
		return
	}

	json.NewEncoder(w).Encode(product)
}

// @Summary List all products
// @Description Get a list of all products
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {array} model.Product
// @Success 204 "No Content"
// @Failure 500 {object} errors.Error
// @Router /products [get]
func (h *ProductHandler) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productRepo.List(r.Context())
	if err != nil {
		json.NewEncoder(w).Encode(errors.NewInternalError("listProducts", err))
		return
	}

	if len(products) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	json.NewEncoder(w).Encode(products)
}
