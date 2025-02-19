package middleware

import (
	"net/http"
	"time"

	"github.com/Napat/golang-testcontainers-demo/pkg/metrics"
)

type MetricsMiddleware struct {
	metrics *metrics.HTTPMetrics
}

func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics.NewHTTPMetrics(),
	}
}

func (m *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response writer wrapper to capture status code
		rw := NewResponseWriter(w)

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics after handler returns
		duration := time.Since(start).Seconds()
		m.metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		m.metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, rw.Status()).Inc()
	})
}

// ResponseWriter wrapper to capture status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Status() string {
	return http.StatusText(rw.statusCode)
}
