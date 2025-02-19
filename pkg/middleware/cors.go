package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig holds configuration for CORS middleware
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		MaxAge:         86400, // 24 hours
	}
}

// CORS returns a middleware that handles CORS
func CORS(config *CORSConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			origin := r.Header.Get("Origin")
			if origin != "" {
				for _, allowedOrigin := range config.AllowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
