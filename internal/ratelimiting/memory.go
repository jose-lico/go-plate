package ratelimiting

import "fmt"

type MemoryRateLimiter struct {
}

func NewMemoryRateLimiter() RateLimiter {
	return &MemoryRateLimiter{}
}

func (*MemoryRateLimiter) Allow() (bool, error) {
	fmt.Println("Check rate limit in redis")
	return true, nil
}
