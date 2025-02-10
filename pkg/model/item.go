package model

import "fmt"

type Item struct {
    ProductID   string  `json:"product_id"`
    ProductName string  `json:"product_name"`
    Quantity    int     `json:"quantity"`
    UnitPrice   float64 `json:"unit_price"`
    Subtotal    float64 `json:"subtotal"`
}

func (i *Item) Validate() error {
    if i.ProductID == "" {
        return fmt.Errorf("product ID is required")
    }
    if i.ProductName == "" {
        return fmt.Errorf("product name is required")
    }
    if i.Quantity <= 0 {
        return fmt.Errorf("quantity must be positive")
    }
    if i.UnitPrice < 0 {
        return fmt.Errorf("unit price must be non-negative")
    }
    // Verify subtotal calculation
    expectedSubtotal := float64(i.Quantity) * i.UnitPrice
    if i.Subtotal != expectedSubtotal {
        return fmt.Errorf("invalid subtotal: expected %.2f, got %.2f", expectedSubtotal, i.Subtotal)
    }
    return nil
}
