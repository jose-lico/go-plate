package ratelimiting

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu           sync.Mutex
	buckets      map[string]*bucket
	rate         float64 // tokens per second
	capacity     float64
	cleanupEvery time.Duration
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

func NewTokenBucket(rate, capacity float64, cleanupInterval time.Duration) *TokenBucket {
	tb := &TokenBucket{
		buckets:      make(map[string]*bucket),
		rate:         rate,
		capacity:     capacity,
		cleanupEvery: cleanupInterval,
	}
	go tb.cleanup()
	return tb
}

func (tb *TokenBucket) Allow(key string) (bool, time.Duration, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b, exists := tb.buckets[key]
	now := time.Now()
	if !exists {
		tb.buckets[key] = &bucket{tokens: tb.capacity - 1, lastRefill: now}
		return true, 0, nil
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(tb.capacity, b.tokens+elapsed*tb.rate)
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true, 0, nil
	}

	retryAfter := time.Duration((1-b.tokens)/tb.rate) * time.Second
	return false, retryAfter, nil
}

func (tb *TokenBucket) cleanup() {
	for range time.Tick(tb.cleanupEvery) {
		tb.mu.Lock()
		for key, bucket := range tb.buckets {
			if time.Since(bucket.lastRefill) > tb.cleanupEvery {
				delete(tb.buckets, key)
			}
		}
		tb.mu.Unlock()
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
