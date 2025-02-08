// internal/model/product.go
package model

import (
	"errors"
	"fmt"
)

type Product struct {
	BaseModel
	Name        string  `json:"name" db:"name"`
	Description string  `json:"description" db:"description"`
	Price       float64 `json:"price" db:"price"`
	SKU         string  `json:"sku" db:"sku"`
	Stock       int     `json:"stock" db:"stock"`
}

// Validate performs basic validation on the product
func (p *Product) Validate() error {
	if p.Name == "" {
		return errors.New("product name is required")
	}

	if p.Price < 0 {
		return fmt.Errorf("invalid price: %.2f", p.Price)
	}

	if p.Stock < 0 {
		return fmt.Errorf("invalid stock quantity: %d", p.Stock)
	}

	if p.SKU == "" {
		return errors.New("SKU is required")
	}

	return nil
}

// CalculateTotal returns the total price for a given quantity
func (p *Product) CalculateTotal(quantity int) (float64, error) {
	if quantity < 0 {
		return 0, fmt.Errorf("invalid quantity: %d", quantity)
	}

	if quantity > p.Stock {
		return 0, fmt.Errorf("insufficient stock: requested %d, available %d", quantity, p.Stock)
	}

	return p.Price * float64(quantity), nil
}

// IsInStock checks if the product is available in the requested quantity
func (p *Product) IsInStock(quantity int) bool {
	return quantity > 0 && p.Stock >= quantity
}

// UpdateStock updates the product stock
func (p *Product) UpdateStock(quantity int) error {
	newStock := p.Stock + quantity
	if newStock < 0 {
		return fmt.Errorf("insufficient stock: current %d, adjustment %d", p.Stock, quantity)
	}
	p.Stock = newStock
	return nil
}
