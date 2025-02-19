package handler

import (
	"net/http"
	"strings"
)

// // ServeMuxAdapter adapts http.ServeMux to our Router interface
// type ServeMuxAdapter struct {
// 	mux *http.ServeMux
// }

// func NewServeMuxAdapter(mux *http.ServeMux) *ServeMuxAdapter {
// 	return &ServeMuxAdapter{mux: mux}
// }

// func (a *ServeMuxAdapter) Handle(pattern string, handler http.Handler) {
// 	a.mux.Handle(pattern, handler)
// }

// func (a *ServeMuxAdapter) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
// 	a.mux.HandleFunc(pattern, handler)
// }

// Router is a simple router interface
type Router interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// Route represents a single route with its method and handler
type Route struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// RouteGroup helps organize routes with a common prefix
type RouteGroup struct {
	prefix string
	routes []Route
}

// NewRouteGroup creates a new route group with prefix
func NewRouteGroup(prefix string) *RouteGroup {
	// Ensure prefix starts with /
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	// Ensure prefix ends with / for consistent path joining
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	return &RouteGroup{
		prefix: prefix,
		routes: []Route{},
	}
}

// Add adds a new route to the group
func (g *RouteGroup) Add(method, pattern string, handler http.HandlerFunc) {
	// Remove leading slash from pattern if present
	pattern = strings.TrimPrefix(pattern, "/")
	g.routes = append(g.routes, Route{
		Method:      method,
		Pattern:     pattern,
		HandlerFunc: handler,
	})
}

// Register registers all routes in the group to the router
func (g *RouteGroup) Register(router Router) {
	for _, route := range g.routes {
		fullPattern := g.prefix + route.Pattern
		// Ensure pattern doesn't have double slashes
		fullPattern = normalizePath(fullPattern)

		router.HandleFunc(fullPattern, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != route.Method {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			route.HandlerFunc(w, r)
		})
	}
}

// normalizePath ensures there are no double slashes in the path
func normalizePath(path string) string {
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}
