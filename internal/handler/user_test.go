package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

type MockProducerRepo struct {
	mock.Mock
}

func (m *MockProducerRepo) SendMessage(topic string, message interface{}) error {
	args := m.Called(topic, message)
	return args.Error(0)
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		input          model.UserCreate
		expectedStatus int
		mockError      error
		setupCache     bool
		setupProducer  bool
	}{
		{
			name: "success",
			input: model.UserCreate{
				Username: "testuser",
				Email:    "test@example.com",
				FullName: "Test User",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			mockError:      nil,
			setupCache:     true,
			setupProducer:  true,
		},
		{
			name: "repository error",
			input: model.UserCreate{
				Username: "testuser",
				Email:    "test@example.com",
			},
			expectedStatus: http.StatusInternalServerError,
			mockError:      assert.AnError,
			setupCache:     false,
			setupProducer:  false,
		},
		{
			name: "duplicate username error",
			input: model.UserCreate{
				Username: "existinguser",
				Email:    "test@example.com",
				FullName: "Test User",
				Password: "password123",
			},
			expectedStatus: http.StatusInternalServerError,
			mockError:      errors.New("Error 1062 (23000): Duplicate entry 'existinguser' for key 'users.username'"),
			setupCache:     false,
			setupProducer:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepo)
			mockCache := new(MockCache)
			mockProducer := new(MockProducerRepo)

			mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(tt.mockError)

			if tt.setupCache {
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}
			if tt.setupProducer {
				mockProducer.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
			}

			handler := handler.NewUserHandler(mockRepo, mockCache, mockProducer)
			routes := handler.GetRoutes()

			// Find the create user route
			var createUserHandler http.HandlerFunc
			for _, route := range routes {
				if route.Method == http.MethodPost && route.Pattern == "/users" {
					createUserHandler = route.Handler
					break
				}
			}

			if createUserHandler == nil {
				t.Fatal("Create user route not found")
			}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			createUserHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockRepo.AssertExpectations(t)
			if tt.setupCache {
				mockCache.AssertExpectations(t)
			}
			if tt.setupProducer {
				mockProducer.AssertExpectations(t)
			}
		})
	}
}

// Additional tests for GetUserByID and GetAllUsers would follow the same pattern...
