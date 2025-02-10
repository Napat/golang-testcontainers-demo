package order

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Napat/golang-testcontainers-demo/internal/repository/repository_order"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/Napat/golang-testcontainers-demo/test/integration"
	"github.com/stretchr/testify/suite"
)

type OrderRepositoryTestSuite struct {
	integration.ElasticTestSuite
	repo *repository_order.OrderRepository
}

func TestOrderRepository(t *testing.T) {
	suite.Run(t, new(OrderRepositoryTestSuite))
}

func (s *OrderRepositoryTestSuite) SetupSuite() {
	s.SetupElasticsearch()
	s.repo = repository_order.NewOrderRepository(s.ESClient)
}

func (s *OrderRepositoryTestSuite) TearDownSuite() {
	s.TearDownElasticsearch()
}

func (s *OrderRepositoryTestSuite) SetupTest() {
	s.setupIndex()
	s.loadTestData()
}

func (s *OrderRepositoryTestSuite) setupIndex() {
	// Delete if exists and create new index
	s.ESClient.Indices.Delete([]string{"orders"})

	// Load and apply template
	template, err := os.ReadFile(filepath.Join("testdata", "order_template.json"))
	s.Require().NoError(err)

	resp, err := s.ESClient.Indices.PutTemplate("orders_template",
		bytes.NewReader(template))
	s.Require().NoError(err)
	resp.Body.Close()

	// Create index
	resp, err = s.ESClient.Indices.Create("orders")
	s.Require().NoError(err)
	resp.Body.Close()
}

func (s *OrderRepositoryTestSuite) loadTestData() {
	// Load seed data
	data, err := os.ReadFile(filepath.Join("testdata", "seed_orders.json"))
	s.Require().NoError(err)

	var orders []model.Order
	err = json.Unmarshal(data, &orders)
	s.Require().NoError(err)

	// Insert seed data
	for _, order := range orders {
		err = s.repo.CreateOrder(context.Background(), &order)
		s.Require().NoError(err)
	}

	// Ensure data is searchable
	_, err = s.ESClient.Indices.Refresh()
	s.Require().NoError(err)
}

func (s *OrderRepositoryTestSuite) TestSearchOrders() {
	ctx := context.Background()

	// Test searching for orders by customer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"customer_id": "cust-001",
			},
		},
	}

	orders, err := s.repo.SearchOrders(ctx, query)
	s.Require().NoError(err)
	s.Require().Len(orders, 2)

	// Test searching by payment method
	query = map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"payment_method": "paypal",
			},
		},
	}

	orders, err = s.repo.SearchOrders(ctx, query)
	s.Require().NoError(err)
	s.Require().Len(orders, 1)
	s.Equal("order-002", orders[0].ID)
}
