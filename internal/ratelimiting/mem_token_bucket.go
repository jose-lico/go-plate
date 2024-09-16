package ratelimiting

import (
	"sync"
	"time"
)

type InMemoryTokenBucket struct {
	mu           sync.Mutex
	buckets      map[string]*bucket
	rate         float64
	capacity     float64
	cleanupEvery time.Duration
}

func NewInMemoryTokenBucket(rate, capacity float64, cleanupInterval time.Duration) RateLimiter {
	tb := &InMemoryTokenBucket{
		buckets:      make(map[string]*bucket),
		rate:         rate,
		capacity:     capacity,
		cleanupEvery: cleanupInterval,
	}
	go tb.cleanup()
	return tb
}

func (tb *InMemoryTokenBucket) Allow(key string) (bool, time.Duration, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b, exists := tb.buckets[key]
	now := time.Now()
	if !exists {
		tb.buckets[key] = &bucket{Tokens: tb.capacity - 1, LastRefill: now}
		return true, 0, nil
	}

	elapsed := now.Sub(b.LastRefill).Seconds()
	b.Tokens = min(tb.capacity, b.Tokens+elapsed*tb.rate)
	b.LastRefill = now

	if b.Tokens >= 1 {
		b.Tokens--
		return true, 0, nil
	}

	retryAfter := time.Duration((1-b.Tokens)/tb.rate) * time.Second
	return false, retryAfter, nil
}

func (tb *InMemoryTokenBucket) cleanup() {
	for range time.Tick(tb.cleanupEvery) {
		tb.mu.Lock()
		for key, bucket := range tb.buckets {
			if time.Since(bucket.LastRefill) > tb.cleanupEvery {
				delete(tb.buckets, key)
			}
		}
		tb.mu.Unlock()
	}
}
