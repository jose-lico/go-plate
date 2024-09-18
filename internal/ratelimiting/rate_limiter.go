package ratelimiting

import (
	"time"
)

type RateLimiter interface {
	Allow(string) (bool, time.Duration, error)
}
