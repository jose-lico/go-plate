package ratelimiting

import (
	"container/list"
	"time"
)

type RateLimiter interface {
	Allow(string) (bool, time.Duration, error)
}

type tokenBucket struct {
	Tokens     float64
	LastRefill time.Time
}

type leakyBucket struct {
	queue      *list.List
	lastLeak   time.Time
	leakPeriod time.Duration
}
