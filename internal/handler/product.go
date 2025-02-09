package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Napat/golang-testcontainers-demo/internal/model"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/product"
)

type ProductHandler struct {
    productRepo *product.ProductRepository
}

func NewProductHandler(productRepo *product.ProductRepository) *ProductHandler {
    return &ProductHandler{
        productRepo: productRepo,
    }
}

func (h *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    switch {
    case r.Method == http.MethodPost && r.URL.Path == "/products":
        h.createProduct(w, r)
    case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/products/"):
        h.getProduct(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
    var product model.Product
    if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := h.productRepo.Create(r.Context(), &product); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request) {
    idStr := strings.TrimPrefix(r.URL.Path, "/products/")
    
    // Convert string ID to int64
    productID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }

    product, err := h.productRepo.GetByID(r.Context(), productID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(product)
}
