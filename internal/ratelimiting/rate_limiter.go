package ratelimiting

import (
	"time"
)

type RateLimiter interface {
	Allow(string) (bool, time.Duration, error)
}

type bucket struct {
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
