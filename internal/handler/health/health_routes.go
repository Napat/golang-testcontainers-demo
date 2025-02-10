package health

import (
	"net/http"
)

func RegisterHealthRoutes(mux *http.ServeMux, healthHandler *HealthHandler) {
	mux.Handle("/health", healthHandler)
	mux.Handle("/health/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	mux.Handle("/health/ready", healthHandler)
}
