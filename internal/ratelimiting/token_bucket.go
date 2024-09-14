package ratelimiting

import (
	"fmt"
	"net/http"
	"time"
)

type TokenBucketAlgo struct {
}

func NewTokenBucketAlgo() Algorithm {
	return &TokenBucketAlgo{}
}

func (*TokenBucketAlgo) CanMakeRequest(store RateLimiter, r *http.Request) (bool, time.Duration, error) {
	fmt.Println("Check rate limit with token algo")

	allow, _ := store.Allow()

	return allow, 0, nil
}
