package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "GET returns degraded status when no connections",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "POST returns method not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHealthHandler(nil, nil, nil, nil, nil)
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse {
				var response HealthStatus
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "degraded", response.Status)
				assert.NotEmpty(t, response.Timestamp)
				assert.Len(t, response.Services, 5)

				// Check all services report as down due to nil connections
				for _, service := range response.Services {
					assert.Equal(t, "down", service.Status)
					assert.Contains(t, service.Message, "not initialized")
				}
			}
		})
	}
}
