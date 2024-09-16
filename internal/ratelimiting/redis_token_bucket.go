package ratelimiting

import (
	"context"
	"encoding/json"
	"time"

	"go-plate/internal/database"

	"github.com/redis/go-redis/v9"
)

type RedisTokenBucket struct {
	redis        database.RedisStore
	rate         float64
	capacity     float64
	cleanupEvery time.Duration
	ctx          context.Context
}

func NewRedisTokenBucket(redis database.RedisStore, rate, capacity float64) RateLimiter {
	return &RedisTokenBucket{
		redis:    redis,
		rate:     rate,
		capacity: capacity,
		ctx:      context.Background(),
	}
}

func (tb *RedisTokenBucket) Allow(key string) (bool, time.Duration, error) {
	bucketState, err := tb.getBucket(key)
	if err != nil {
		return false, 0, err
	}

	now := time.Now()
	elapsed := now.Sub(bucketState.LastRefill).Seconds()

	bucketState.Tokens = min(tb.capacity, bucketState.Tokens+elapsed*tb.rate)
	bucketState.LastRefill = now

	if bucketState.Tokens >= 1 {
		bucketState.Tokens--
		tb.saveBucket(key, bucketState)
		return true, 0, nil
	}

	retryAfter := time.Duration((1-bucketState.Tokens)/tb.rate) * time.Second
	tb.saveBucket(key, bucketState)
	return false, retryAfter, nil
}

func (tb *RedisTokenBucket) getBucket(key string) (*bucket, error) {
	result, err := tb.redis.Get(tb.ctx, "ratelimit:"+key)
	if err == redis.Nil {
		return &bucket{
			Tokens:     tb.capacity - 1,
			LastRefill: time.Now(),
		}, nil
	} else if err != nil {
		return nil, err
	}

	var b bucket
	err = json.Unmarshal([]byte(result), &b)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (tb *RedisTokenBucket) saveBucket(key string, b *bucket) error {
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}

	return tb.redis.Set(tb.ctx, "ratelimit:"+key, data, tb.cleanupEvery)
}
