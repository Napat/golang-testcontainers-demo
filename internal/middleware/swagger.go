package middleware

import (
	"net/http"

	"github.com/swaggo/http-swagger"
)

func Swagger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/swagger/*" {
				httpSwagger.Handler(
					httpSwagger.URL("/swagger/doc.json"),
				).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
