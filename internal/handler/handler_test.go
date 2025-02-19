package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Napat/golang-testcontainers-demo/internal/handler/health"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) CreateOrder(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) SearchOrders(ctx context.Context, params map[string]interface{}) ([]model.Order, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Order), args.Error(1)
}

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

func (m *MockProductRepo) GetAll(ctx context.Context) ([]*model.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Product), args.Error(1)
}

func (m *MockProductRepo) Update(ctx context.Context, product *model.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetAll(ctx context.Context) ([]*model.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.User), args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

type MockMessageProducer struct {
	mock.Mock
}

func (m *MockMessageProducer) SendMessage(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func TestOrderHandler(t *testing.T) {
	mockRepo := new(MockOrderRepo)
	mockRepo.On("SearchOrders", mock.Anything, mock.Anything).Return([]model.Order{}, nil)

	handler := NewOrderHandler(mockRepo)
	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	handler.ListOrders(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status NoContent; got %v", w.Code)
	}
	mockRepo.AssertExpectations(t)
}

func TestProductHandler(t *testing.T) {
	mockRepo := new(MockProductRepo)
	mockRepo.On("GetAll", mock.Anything).Return([]*model.Product{}, nil)

	handler := NewProductHandler(mockRepo)
	req := httptest.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()

	handler.getAllProducts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}
	mockRepo.AssertExpectations(t)
}

func TestUserHandler(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockCache := new(MockCache)
	mockProducer := new(MockMessageProducer)

	mockRepo.On("GetAll", mock.Anything).Return([]*model.User{}, nil)

	handler := NewUserHandler(mockRepo, mockCache, mockProducer)
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	handler.getAllUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}
	mockRepo.AssertExpectations(t)
}

func TestHealthHandler(t *testing.T) {
	handler := health.NewHealthHandler(nil, nil, nil, nil, nil)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}
}
