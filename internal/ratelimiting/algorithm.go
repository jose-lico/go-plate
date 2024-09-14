package ratelimiting

import (
	"net/http"
	"time"
)

type Algorithm interface {
	CanMakeRequest(RateLimiter, *http.Request) (bool, time.Duration, error)
}

var (
	TokenBucket Algorithm
	LeakyBucket Algorithm
)
