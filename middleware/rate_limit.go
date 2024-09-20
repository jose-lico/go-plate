package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/jose-lico/go-plate/ratelimiting"
)

func RateLimitMiddleware(limiter ratelimiting.RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, retryAfter, err := limiter.Allow(r.Header.Get("X-Real-IP"))

			if err != nil {
				log.Printf("[ERROR] Error rate limiting request: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", time.Now().Add(retryAfter).Format(time.RFC1123))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
