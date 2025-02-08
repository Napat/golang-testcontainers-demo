package model

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`  // "-" means this field won't be included in JSON
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at,omitempty"`
}

// UserStatus represents the possible states of a user account
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusBanned   = "banned"
)

// NewUser creates a new user with default values
func NewUser(username, email, password string) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Email:     email,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Validate performs validation on user data
func (u *User) Validate() error {
	if err := u.validateUsername(); err != nil {
		return err
	}

	if err := u.validateEmail(); err != nil {
		return err
	}

	if u.Status != StatusActive && u.Status != StatusInactive && u.Status != StatusBanned {
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
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
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

// Ban sets the user's status to banned
func (u *User) Ban() {
	u.Status = StatusBanned
	u.UpdatedAt = time.Now()
}

// ToPublicUser returns a copy of the user with sensitive information removed
func (u *User) ToPublicUser() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"username":   u.Username,
		"email":      u.Email,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"status":     u.Status,
		"created_at": u.CreatedAt,
	}
}
