package model

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

// User represents a user in the system
// @Description User account information
type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	RowNumber int        `json:"row_number,omitempty" db:"-"`
	Username  string     `json:"username" db:"username"`
	Email     string     `json:"email" db:"email"`
	FullName  string     `json:"full_name" db:"full_name"`
	Password  string     `json:"password,omitempty" db:"password"`
	Status    UserStatus `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Version   int        `json:"version" db:"version"`
}

// UserCreate represents user creation request
// @Description User creation request body
type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// UserUpdate represents user update request
// @Description User update request body
type UserUpdate struct {
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	FullName string `json:"full_name,omitempty"`
	Status   string `json:"status,omitempty" binding:"omitempty,oneof=active inactive suspended"`
}

// NewUser creates a new user with default values
func NewUser(username, email, password string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.Must(uuid.NewV7()),
		Username:  username,
		Email:     email,
		Password:  password,
		Status:    StatusActive, // Explicitly set as StatusActive
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

// SetPassword sets the user's password directly
func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	u.Password = password
	return nil
}

// CheckPassword verifies if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	return u.Password == password
}

// Validate performs validation on user data
func (u *User) Validate() error {
	if err := u.validateUsername(); err != nil {
		return err
	}

	if err := u.validateEmail(); err != nil {
		return err
	}

	if u.Status != StatusActive && u.Status != StatusInactive && u.Status != StatusSuspended {
		return errors.New("invalid user status")
	}

	return nil
}

// validateUsername checks if the username is valid
func (u *User) validateUsername() error {
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if len(u.Username) > 50 {
		return errors.New("username must not exceed 50 characters")
	}

	// Check if username contains only allowed characters
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", u.Username)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("username can only contain letters, numbers, underscores, and hyphens")
	}

	return nil
}

// validateEmail checks if the email is valid
func (u *User) validateEmail() error {
	if u.Email == "" {
		return errors.New("email is required")
	}

	// Remove leading/trailing whitespace
	u.Email = strings.TrimSpace(u.Email)

	// Parse email address
	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return errors.New("invalid email format")
	}

	return nil
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return strings.TrimSpace(u.FullName)
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// Deactivate sets the user's status to inactive
func (u *User) Deactivate() {
	u.Status = StatusInactive
	u.UpdatedAt = time.Now()
}

// Activate sets the user's status to active
func (u *User) Activate() {
	u.Status = StatusActive
	u.UpdatedAt = time.Now()
}

// Ban sets the user's status to suspended
func (u *User) Ban() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now()
}

// ToPublicUser returns a copy of the user with sensitive information removed
func (u *User) ToPublicUser() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"username":   u.Username,
		"email":      u.Email,
		"full_name":  u.FullName,
		"status":     u.Status,
		"created_at": u.CreatedAt,
	}
}
