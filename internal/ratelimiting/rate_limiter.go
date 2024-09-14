package ratelimiting

type RateLimiter interface {
	Allow() (bool, error)
}

var (
	Redis  RateLimiter
	Memory RateLimiter
)
