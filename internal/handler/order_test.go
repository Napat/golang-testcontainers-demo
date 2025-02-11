package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) CreateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) SearchOrders(ctx context.Context, query map[string]interface{}) ([]model.Order, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]model.Order), args.Error(1)
}

func TestOrderHandler_CreateOrder(t *testing.T) {
	tests := []struct {
		name           string
		input          model.Order
		expectedStatus int
		mockError      error
		setupMock      bool
	}{
		{
			name: "success",
			input: model.Order{
				ID:         "order-1", // Required field
				CustomerID: "1",
				Status:     "pending",
				Items: []model.Item{
					{
						ProductID:   "1",
						ProductName: "Test Product", // Required field
						Quantity:    1,
						UnitPrice:   10.00,
						Subtotal:    10.00,
					},
				},
				Total:         10.00,  // Required field
				PaymentMethod: "card", // Required field
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			expectedStatus: http.StatusCreated,
			mockError:      nil,
			setupMock:      true,
		},
		{
			name: "validation error",
			input: model.Order{
				CustomerID: "1",
				// Missing required fields to trigger validation error
			},
			expectedStatus: http.StatusUnprocessableEntity,
			setupMock:      false,
		},
		{
			name: "repository error",
			input: model.Order{
				ID:         "order-2",
				CustomerID: "1",
				Status:     "pending",
				Items: []model.Item{
					{
						ProductID:   "1",
						ProductName: "Test Product",
						Quantity:    1,
						UnitPrice:   10.00,
						Subtotal:    10.00,
					},
				},
				Total:         10.00,
				PaymentMethod: "card",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			expectedStatus: http.StatusInternalServerError,
			mockError:      assert.AnError,
			setupMock:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockOrderRepo)
			if tt.setupMock {
				mockRepo.On("CreateOrder", mock.Anything, mock.AnythingOfType("*model.Order")).Return(tt.mockError)
			}

			handler := handler.NewOrderHandler(mockRepo)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.setupMock {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

// Additional tests for SearchOrders and ListOrders would follow the same pattern...
