package routes

import "net/http"

// Route represents a single route with its HTTP method, pattern and handler
type Route struct {
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

// Handler interface for all handlers that can provide their routes
type Handler interface {
	GetRoutes() []Route
}
