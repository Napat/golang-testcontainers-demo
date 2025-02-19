package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductRepo struct {
	mock.Mock
}

func (m *MockProductRepo) Create(ctx context.Context, product *model.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepo) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockProductRepo) List(ctx context.Context) ([]*model.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Product), args.Error(1)
}

func (m *MockProductRepo) GetAll(ctx context.Context) ([]*model.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Product), args.Error(1)
}

func (m *MockProductRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepo) Update(ctx context.Context, product *model.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func TestProductHandler_CreateProduct(t *testing.T) {
	tests := []struct {
		name           string
		input          model.Product
		expectedStatus int
		mockError      error
	}{
		{
			name: "success",
			input: model.Product{
				Name:        "Test Product",
				Price:       9.99,
				Description: "Product description 1",
			},
			expectedStatus: http.StatusCreated,
			mockError:      nil,
		},
		{
			name: "repository error",
			input: model.Product{
				Name:  "Test Product",
				Price: 9.99,
			},
			expectedStatus: http.StatusInternalServerError,
			mockError:      assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProductRepo)
			mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Product")).Return(tt.mockError)

			handler := handler.NewProductHandler(mockRepo)
			routes := handler.GetRoutes()

			// Find the create product route
			var createProductHandler http.HandlerFunc
			for _, route := range routes {
				if route.Method == http.MethodPost && route.Pattern == "/products" {
					createProductHandler = route.Handler
					break
				}
			}

			if createProductHandler == nil {
				t.Fatal("Create product route not found")
			}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			createProductHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// Additional tests for GetProduct and ListProducts would follow the same pattern...
