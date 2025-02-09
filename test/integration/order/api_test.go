package order

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/internal/model"
	"github.com/Napat/golang-testcontainers-demo/internal/repository/order"
	"github.com/Napat/golang-testcontainers-demo/test/integration"
	"github.com/stretchr/testify/suite"
)

type OrderAPITestSuite struct {
	integration.ElasticTestSuite
	repo    *order.OrderRepository
	handler http.Handler
}

// TestOrderAPI runs the OrderAPITestSuite using the testify/suite package.
func TestOrderAPI(t *testing.T) {
	suite.Run(t, new(OrderAPITestSuite))
}

// SetupSuite sets up the necessary resources and components for the
// OrderAPITestSuite. It initializes the Elasticsearch client, creates a new
// OrderRepository, and configures the HTTP router with order-related handlers.

func (s *OrderAPITestSuite) SetupSuite() {
	s.SetupElasticsearch()
	s.repo = order.NewOrderRepository(s.ESClient)

	router := http.NewServeMux()
	orderHandler := handler.NewOrderHandler(s.repo)
	router.HandleFunc("/orders", orderHandler.ServeHTTP)
	router.HandleFunc("/orders/search", orderHandler.ServeHTTP)
	s.handler = router
}

// SetupTest sets up a fresh index for each test. It deletes the existing
// index, creates a new one with the order template, and waits for the index
// to be ready.
func (s *OrderAPITestSuite) SetupTest() {
	// Clear existing index
	_, err := s.ESClient.Indices.Delete([]string{"orders"})
	s.Require().NoError(err)

	// Create new index with template
	templateFile := filepath.Join("testdata", "order_template.json")
	template, err := os.ReadFile(templateFile)
	s.Require().NoError(err)

	resp, err := s.ESClient.Indices.PutTemplate("orders_template",
		bytes.NewReader(template))
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Create index
	resp, err = s.ESClient.Indices.Create("orders")
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Wait for index to be ready
	_, err = s.ESClient.Indices.Refresh()
	s.Require().NoError(err)
}

// TestCreateOrderAPI tests the /orders endpoint for creating a new order.
// It sends a POST request with a valid order payload and checks that the
// response status code is 201 Created and the response body contains the
// newly-created order data.
func (s *OrderAPITestSuite) TestCreateOrderAPI() {
	order := model.Order{
		ID:            "test-order-1",
		CustomerID:    "cust-1",
		Status:        "pending",
		Total:         299.99,
		PaymentMethod: "credit_card",
		Items: []model.Item{
			{
				ProductID:   "prod-1",
				ProductName: "Test Product",
				Quantity:    1,
				UnitPrice:   299.99,
				Subtotal:    299.99,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	body, err := json.Marshal(order)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.handler.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)

	var createdOrder model.Order
	err = json.NewDecoder(rec.Body).Decode(&createdOrder)
	s.Require().NoError(err)
	s.Equal(order.ID, createdOrder.ID)
}

// TestSearchOrdersAPI tests the /orders/search endpoint for searching for
// orders. It first creates a test order, then sends a GET request with a valid
// query payload and checks that the response status code is 200 OK and the
// response body contains the expected order data.
func (s *OrderAPITestSuite) TestSearchOrdersAPI() {
	// Create a test order
	order := model.Order{
		ID:            "test-order-2",
		CustomerID:    "cust-2",
		Status:        "pending",
		Total:         159.98,
		PaymentMethod: "paypal",
		Items: []model.Item{
			{
				ProductID:   "prod-2",
				ProductName: "Test Product",
				Quantity:    2,
				UnitPrice:   79.99,
				Subtotal:    159.98,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := s.repo.CreateOrder(context.Background(), &order)
	s.Require().NoError(err)

	// Ensure the index is refreshed
	_, err = s.ESClient.Indices.Refresh()
	s.Require().NoError(err)

	// Then search for it
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"customer_id": "cust-2",
			},
		},
	}

	body, err := json.Marshal(query)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/orders/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.handler.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var orders []model.Order
	err = json.NewDecoder(rec.Body).Decode(&orders)
	s.Require().NoError(err)
	s.Require().Len(orders, 1)
	s.Equal(order.ID, orders[0].ID)
}
