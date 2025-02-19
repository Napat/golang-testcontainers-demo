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
	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_order"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/Napat/golang-testcontainers-demo/pkg/testhelper"
	"github.com/Napat/golang-testcontainers-demo/test/integration"
	"github.com/stretchr/testify/suite"
)

type OrderAPITestSuite struct {
	integration.ElasticTestSuite
	repo    *repository_order.OrderRepository
	handler http.Handler
}

// TestIntegrationOrderAPI runs the OrderAPITestSuite using the testify/suite package.
func TestIntegrationOrderAPI(t *testing.T) {
	testhelper.SkipIfShort(t)
	t.Parallel()
	suite.Run(t, new(OrderAPITestSuite))
}

// SetupSuite sets up the necessary resources and components for the
// OrderAPITestSuite. It initializes the Elasticsearch client, creates a new
// OrderRepository, and configures the HTTP router with order-related handlers.

func (s *OrderAPITestSuite) SetupSuite() {
	s.SetupElasticsearch()
	s.repo = repository_order.NewOrderRepository(s.ESClient)

	// Create a new handler
	orderHandler := handler.NewOrderHandler(s.repo)

	// Create a router that forwards all /api/v1/orders* requests to the order handler
	router := http.NewServeMux()
	router.Handle("/api/v1/orders/", orderHandler)
	router.Handle("/api/v1/orders", orderHandler)
	s.handler = router
}

// SetupTest sets up a fresh index for each test. It deletes the existing
// index, creates a new one with the order template, and waits for the index
// to be ready.
func (s *OrderAPITestSuite) SetupTest() {
	// Delete index if exists
	exists, err := s.ESClient.Indices.Exists([]string{"orders"})
	s.Require().NoError(err)
	if !exists.IsError() {
		_, err = s.ESClient.Indices.Delete([]string{"orders"})
		s.Require().NoError(err)
	}

	// Delete template if exists
	_, err = s.ESClient.Indices.DeleteTemplate("orders_template")
	s.Require().NoError(err)

	// Create new template
	templateFile := filepath.Join("testdata", "order_template.json")
	template, err := os.ReadFile(templateFile)
	s.Require().NoError(err)

	resp, err := s.ESClient.Indices.PutTemplate("orders_template",
		bytes.NewReader(template),
		s.ESClient.Indices.PutTemplate.WithCreate(true))
	s.Require().NoError(err)
	s.Require().False(resp.IsError(), "Failed to put template: %s", resp.String())
	resp.Body.Close()

	// Create index
	resp, err = s.ESClient.Indices.Create("orders",
		s.ESClient.Indices.Create.WithWaitForActiveShards("1"))
	s.Require().NoError(err)
	s.Require().False(resp.IsError(), "Failed to create index: %s", resp.String())
	resp.Body.Close()

	// Verify index settings and mappings
	resp, err = s.ESClient.Indices.GetMapping(s.ESClient.Indices.GetMapping.WithIndex("orders"))
	s.Require().NoError(err)
	s.Require().False(resp.IsError(), "Failed to get mapping: %s", resp.String())

	var mapping map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&mapping)
	s.Require().NoError(err)
	resp.Body.Close()

	// Verify customer_id field is keyword type
	ordersMapping := mapping["orders"].(map[string]interface{})
	mappings := ordersMapping["mappings"].(map[string]interface{})
	properties := mappings["properties"].(map[string]interface{})
	customerID := properties["customer_id"].(map[string]interface{})
	s.Require().Equal("keyword", customerID["type"], "customer_id should be keyword type")

	// Additional wait to ensure cluster is ready
	time.Sleep(1 * time.Second)
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

	// Change URL to include /api/v1 prefix
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.handler.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)

	var createdOrder model.Order
	err = json.NewDecoder(rec.Body).Decode(&createdOrder)
	s.Require().NoError(err)
	s.Equal(order.ID, createdOrder.ID)

	// Wait for indexing
	_, err = s.ESClient.Indices.Refresh(
		s.ESClient.Indices.Refresh.WithIndex("orders"),
	)
	s.Require().NoError(err)
	time.Sleep(1 * time.Second)
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

	// Verify document is indexed
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		s.T().Fatalf("Error encoding query: %v", err)
	}

	searchResp, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(context.Background()),
		s.ESClient.Search.WithIndex("orders"),
		s.ESClient.Search.WithBody(&buf),
		s.ESClient.Search.WithTrackTotalHits(true),
	)
	s.Require().NoError(err)
	defer searchResp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(searchResp.Body).Decode(&result); err != nil {
		s.T().Fatalf("Error parsing the response body: %s", err)
	}

	hits := result["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})["value"].(float64)
	s.Require().Equal(float64(1), total, "Expected 1 document in index")

	// Verify index count and refresh before search endpoint call
	// (manual verification already done above)
	_, err = s.ESClient.Indices.Refresh(s.ESClient.Indices.Refresh.WithIndex("orders"))
	s.Require().NoError(err)
	time.Sleep(500 * time.Millisecond)

	// Then search for it
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/search?customer_id=cust-2", nil)
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

func (s *OrderAPITestSuite) TestSimpleSearchAPI() {
	// Create test order
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

	err := s.repo.CreateOrder(context.Background(), &order)
	s.Require().NoError(err)

	// Verify document is indexed
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		s.T().Fatalf("Error encoding query: %v", err)
	}

	searchResp, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(context.Background()),
		s.ESClient.Search.WithIndex("orders"),
		s.ESClient.Search.WithBody(&buf),
		s.ESClient.Search.WithTrackTotalHits(true),
	)
	s.Require().NoError(err)
	defer searchResp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(searchResp.Body).Decode(&result); err != nil {
		s.T().Fatalf("Error parsing the response body: %s", err)
	}

	hits := result["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})["value"].(float64)
	s.Require().Equal(float64(1), total, "Expected 1 document in index")

	// Verify index count and refresh before simple search call
	_, err = s.ESClient.Indices.Refresh(s.ESClient.Indices.Refresh.WithIndex("orders"))
	s.Require().NoError(err)
	time.Sleep(500 * time.Millisecond)

	// Test simple search
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/simple-search?q=cust-1", nil)
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response []model.Order
	err = json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)
	s.Require().Len(response, 1)
	s.Equal("cust-1", response[0].CustomerID)
}
