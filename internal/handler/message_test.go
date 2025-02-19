package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Napat/golang-testcontainers-demo/internal/handler"
	"github.com/Napat/golang-testcontainers-demo/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessage(topic string, message interface{}) error {
	args := m.Called(topic, message)
	return args.Error(0)
}

func TestMessageHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    model.MessageRequest
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			requestBody:    model.MessageRequest{Content: "test message"},
			mockError:      nil,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "producer error",
			method:         http.MethodPost,
			requestBody:    model.MessageRequest{Content: "test message"},
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProducer := new(MockProducer)
			if tt.method == http.MethodPost && tt.mockError != nil {
				mockProducer.On("SendMessage", "message", tt.requestBody).Return(tt.mockError)
			} else if tt.method == http.MethodPost {
				mockProducer.On("SendMessage", "message", tt.requestBody).Return(nil)
			}

			h := handler.NewMessageHandler(mockProducer)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/messages", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockProducer.AssertExpectations(t)
		})
	}
}
