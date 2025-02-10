package errors

import (
	"fmt"
	"net/http"
)

// Error types
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Op      string `json:"op,omitempty"` // Operation that failed
	Err     error  `json:"-"`            // Underlying error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Error constructors
func NewNotFound(op string, entity string, id interface{}) *Error {
	return &Error{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s with id %v not found", entity, id),
		Op:      op,
	}
}

func NewBadRequest(op string, message string) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Message: message,
		Op:      op,
	}
}

func NewInternalError(op string, err error) *Error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
		Op:      op,
		Err:     err,
	}
}

func NewValidationError(op string, message string) *Error {
	return &Error{
		Code:    http.StatusUnprocessableEntity,
		Message: message,
		Op:      op,
	}
}
