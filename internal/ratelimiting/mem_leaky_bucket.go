package ratelimiting

import (
	"container/list"
	"log"
	"sync"
	"time"
)

type InMemoryLeakyBucket struct {
	mu           sync.Mutex
	buckets      map[string]*leakyBucket
	rate         float64
	capacity     int
	cleanupEvery time.Duration
}

func NewInMemoryLeakyBucket(rate float64, capacity int, cleanupInterval time.Duration) RateLimiter {
	if rate <= 0 || capacity <= 0 || cleanupInterval <= 0 {
		log.Fatalf("[FATAL] Invalid parameters for InMemoryLeakyBucket")
	}

	lb := &InMemoryLeakyBucket{
		buckets:      make(map[string]*leakyBucket),
		rate:         rate,
		capacity:     capacity,
		cleanupEvery: cleanupInterval,
	}
	go lb.cleanup()
	return lb
}

func (lb *InMemoryLeakyBucket) Allow(key string) (bool, time.Duration, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	b, exists := lb.buckets[key]
	now := time.Now()

	if !exists {
		b = &leakyBucket{
			queue:      list.New(),
			lastLeak:   now,
			leakPeriod: time.Duration(float64(time.Second) / lb.rate),
		}
		lb.buckets[key] = b
	}

	lb.leak(b, now)

	if b.queue.Len() < lb.capacity {
		b.queue.PushBack(now)
		return true, 0, nil
	} else {
		earliest := b.queue.Front().Value.(time.Time)
		waitTime := b.leakPeriod - now.Sub(b.lastLeak)
		retryAfter := earliest.Add(time.Duration(lb.capacity) * b.leakPeriod).Sub(now)
		if retryAfter < 0 {
			retryAfter = waitTime
		}
		return false, retryAfter, nil
	}
}

func (lb *InMemoryLeakyBucket) leak(b *leakyBucket, now time.Time) {
	elapsed := now.Sub(b.lastLeak)
	leaks := int(elapsed / b.leakPeriod)
	if leaks <= 0 {
		return
	}

	for i := 0; i < leaks && b.queue.Len() > 0; i++ {
		b.queue.Remove(b.queue.Front())
		b.lastLeak = b.lastLeak.Add(b.leakPeriod)
	}

	if b.lastLeak.After(now) {
		b.lastLeak = now
	}
}

func (lb *InMemoryLeakyBucket) cleanup() {
	ticker := time.NewTicker(lb.cleanupEvery)
	defer ticker.Stop()
	for range ticker.C {
		lb.mu.Lock()
		for key, b := range lb.buckets {
			if time.Since(b.lastLeak) > lb.cleanupEvery {
				delete(lb.buckets, key)
			}
		}
		lb.mu.Unlock()
	}
}
