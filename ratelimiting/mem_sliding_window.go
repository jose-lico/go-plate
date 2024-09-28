package ratelimiting

import (
	"log"
	"sync"
	"time"
)

type slidingWindow struct {
	counts        []int
	lastUpdate    time.Time
	windowSize    time.Duration
	subWindowSize time.Duration
	numSubWindows int
}

type InMemorySlidingWindowCounter struct {
	mu            sync.Mutex
	windows       map[string]*slidingWindow
	rate          int
	numSubWindows int
	windowSize    time.Duration
	subWindowSize time.Duration
	cleanupEvery  time.Duration
}

func NewInMemorySlidingWindowCounter(rate int, windowSize, subWindowSize, cleanupInterval time.Duration) RateLimiter {
	if rate <= 0 || windowSize <= 0 || subWindowSize <= 0 || cleanupInterval <= 0 {
		log.Fatalf("[FATAL] Invalid parameters for InMemorySlidingWindowCounter")
	}

	numSubWindows := int(windowSize / subWindowSize)
	if numSubWindows <= 0 {
		log.Fatalf("[FATAL] windowSize must be greater than subWindowSize")
	}

	sw := &InMemorySlidingWindowCounter{
		windows:       make(map[string]*slidingWindow),
		rate:          rate,
		windowSize:    windowSize,
		subWindowSize: subWindowSize,
		numSubWindows: numSubWindows,
		cleanupEvery:  cleanupInterval,
	}

	go sw.cleanup()
	return sw
}

func (swc *InMemorySlidingWindowCounter) Allow(key string) (bool, time.Duration, error) {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()

	window, exists := swc.windows[key]
	if !exists {
		window = &slidingWindow{
			counts:        make([]int, swc.numSubWindows),
			lastUpdate:    now,
			windowSize:    swc.windowSize,
			subWindowSize: swc.subWindowSize,
			numSubWindows: swc.numSubWindows,
		}
		currentSubWindow := swc.getSubWindowIndex(now)
		window.counts[currentSubWindow] = 1
		swc.windows[key] = window
		return true, 0, nil
	}

	elapsed := now.Sub(window.lastUpdate)
	elapsedSubWindows := int(elapsed / swc.subWindowSize)

	if elapsedSubWindows >= swc.numSubWindows {
		for i := 0; i < swc.numSubWindows; i++ {
			window.counts[i] = 0
		}
		window.counts[swc.getSubWindowIndex(now)] = 1
		window.lastUpdate = now
		return true, 0, nil
	}

	for i := 0; i < elapsedSubWindows; i++ {
		oldestSubWindow := (swc.getSubWindowIndex(window.lastUpdate) + 1) % swc.numSubWindows
		window.counts[oldestSubWindow] = 0
		window.lastUpdate = window.lastUpdate.Add(swc.subWindowSize)
	}

	currentSubWindow := swc.getSubWindowIndex(now)

	total := 0
	for _, count := range window.counts {
		total += count
	}

	if total >= swc.rate {
		for i := 0; i < swc.numSubWindows; i++ {
			idx := (currentSubWindow + i + 1) % swc.numSubWindows
			if window.counts[idx] > 0 {
				retryAfter := window.lastUpdate.Add(time.Duration(i+1) * swc.subWindowSize).Sub(now)
				return false, retryAfter, nil
			}
		}
		return false, swc.windowSize, nil
	}

	window.counts[currentSubWindow]++

	return true, 0, nil
}

func (swc *InMemorySlidingWindowCounter) getSubWindowIndex(t time.Time) int {
	return int(t.UnixNano()/swc.subWindowSize.Nanoseconds()) % swc.numSubWindows
}

func (swc *InMemorySlidingWindowCounter) cleanup() {
	ticker := time.NewTicker(swc.cleanupEvery)
	defer ticker.Stop()

	for range ticker.C {
		swc.mu.Lock()
		now := time.Now()
		for key, window := range swc.windows {
			if now.Sub(window.lastUpdate) > swc.windowSize {
				delete(swc.windows, key)
			}
		}
		swc.mu.Unlock()
	}
}
