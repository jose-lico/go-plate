package middleware

import (
	"net/http"

	"go-plate/internal/ratelimiting"
)

func RateLimitMiddleware(algo ratelimiting.Algorithm, rateLimiter ratelimiting.RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, _, err := algo.CanMakeRequest(rateLimiter, r)

			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", "5000")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
