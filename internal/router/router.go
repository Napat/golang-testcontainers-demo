package router

import (
	"net/http"

	"github.com/Napat/golang-testcontainers-demo/internal/config"
	"github.com/Napat/golang-testcontainers-demo/internal/handler/health"
	"github.com/Napat/golang-testcontainers-demo/pkg/middleware"
	"github.com/Napat/golang-testcontainers-demo/pkg/routes"
)

func Setup(
	userHandler routes.Handler,
	productHandler routes.Handler,
	orderHandler routes.Handler,
	messageHandler routes.Handler,
	healthHandler *health.HealthHandler,
	cfg *config.Config,
) http.Handler {
	mux := http.NewServeMux()
	apiMux := http.NewServeMux()

	// Create middleware chain
	var middlewares []func(http.Handler) http.Handler

	// Add CORS middleware if enabled
	if cfg.Server.CORS.Enabled {
		middlewares = append(middlewares, middleware.CORS(&middleware.CORSConfig{
			AllowedOrigins: cfg.Server.CORS.AllowedOrigins,
			AllowedMethods: cfg.Server.CORS.AllowedMethods,
			AllowedHeaders: cfg.Server.CORS.AllowedHeaders,
			MaxAge:         cfg.Server.CORS.MaxAge,
		}))
	}

	// Add other middlewares
	middlewares = append(middlewares,
		middleware.NewMetricsMiddleware().Handler,
		middleware.Tracing(cfg.Tracing.ServiceName),
		middleware.Profiling(),
		middleware.ErrorHandler,
	)

	// Register health check routes on the root mux first, before any middleware
	healthHandler.RegisterRoutes(mux)

	// Handle API routes
	routeHandlers := make(map[string]map[string]http.HandlerFunc)
	allRoutes := make([]routes.Route, 0)
	allRoutes = append(allRoutes, userHandler.GetRoutes()...)
	allRoutes = append(allRoutes, productHandler.GetRoutes()...)
	allRoutes = append(allRoutes, orderHandler.GetRoutes()...)
	allRoutes = append(allRoutes, messageHandler.GetRoutes()...)

	for _, route := range allRoutes {
		if routeHandlers[route.Pattern] == nil {
			routeHandlers[route.Pattern] = make(map[string]http.HandlerFunc)
		}
		routeHandlers[route.Pattern][route.Method] = route.Handler
	}

	for pattern, handlers := range routeHandlers {
		apiMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if handler, ok := handlers[r.Method]; ok {
				handler(w, r)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		})
	}

	// Apply middleware chain to API routes only
	handler := middleware.Chain(middlewares...)(apiMux)

	// Mount API routes under /api/v1 after health check routes
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", handler))

	return mux
}
