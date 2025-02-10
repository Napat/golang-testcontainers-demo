package model

import (
	"fmt"
	"time"
)

type Order struct {
    ID            string    `json:"id"`
    CustomerID    string    `json:"customer_id"`
    Status        string    `json:"status"`
    Total         float64   `json:"total"`
    PaymentMethod string    `json:"payment_method"`
    Items         []Item    `json:"items"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

func (o *Order) Validate() error {
	if o.ID == "" {
		return fmt.Errorf("order ID is required")
	}
	if o.CustomerID == "" {
		return fmt.Errorf("customer ID is required")
	}
	if o.Total < 0 {
		return fmt.Errorf("total must be non-negative")
	}
	if len(o.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}
	for i, item := range o.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("invalid item at index %d: %w", i, err)
		}
	}
	return nil
}
