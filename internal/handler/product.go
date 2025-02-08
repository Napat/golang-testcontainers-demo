package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Napat/golang-testcontainers-demo/internal/model"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/product"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
    productRepo *product.ProductRepository
}

func NewProductHandler(productRepo *product.ProductRepository) *ProductHandler {
    return &ProductHandler{
        productRepo: productRepo,
    }
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
    var product model.Product
    ctx := r.Context()
    if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := h.productRepo.Create(ctx, &product); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    id, err := strconv.ParseInt(vars["id"], 10, 64)
    if err != nil {
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }

    product, err := h.productRepo.GetByID(ctx, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}
