package user

import (
	"encoding/json"
	"net/http"

	"github.com/Napat/golang-testcontainers-demo/internal/model"
)

// @title Testcontainers Demo API
// @version 1.0
// @description User management endpoints
// @host localhost:8080
// @BasePath /

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	// ...existing code...
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.UserCreate true "User creation request"
// @Success 201 {object} model.User
// @Failure 400 {object} model.Error
// @Failure 500 {object} model.Error
// @Router /users [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var req model.UserCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// ...existing error handling...
	}
	// ...existing code...
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user details by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} model.User
// @Failure 404 {object} model.Error
// @Failure 500 {object} model.Error
// @Router /users/{id} [get]
func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	// ...existing code...
}

// ...existing code...
