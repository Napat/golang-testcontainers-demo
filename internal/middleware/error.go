package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Napat/golang-testcontainers-demo/internal/errors"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				response := ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func WriteError(w http.ResponseWriter, err error) {
	var response ErrorResponse

	switch e := err.(type) {
	case *errors.Error:
		response = ErrorResponse{
			Code:    e.Code,
			Message: e.Message,
		}
	default:
		response = ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		}
		log.Printf("unexpected error: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	json.NewEncoder(w).Encode(response)
}
