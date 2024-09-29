package ratelimiting

import (
	"log"
	"sync"
	"time"
)

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

type InMemoryTokenBucket struct {
	mu           sync.Mutex
	buckets      map[string]*tokenBucket
	rate         float64
	capacity     float64
	cleanupEvery time.Duration
}

func NewInMemoryTokenBucket(rate, capacity float64, cleanupInterval time.Duration) RateLimiter {
	if rate <= 0 || capacity <= 0 || cleanupInterval <= 0 {
		log.Fatalf("[FATAL] Invalid parameters for InMemoryTokenBucket")
	}

	tb := &InMemoryTokenBucket{
		buckets:      make(map[string]*tokenBucket),
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
		tb.buckets[key] = &tokenBucket{tokens: tb.capacity - 1, lastRefill: now}
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

func (tb *InMemoryTokenBucket) cleanup() {
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
