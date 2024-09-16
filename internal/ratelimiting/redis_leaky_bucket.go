package ratelimiting

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-plate/internal/database"

	"github.com/redis/go-redis/v9"
)

type RedisLeakyBucket struct {
	redis         database.RedisStore
	rate          float64
	capacity      int
	keyExpiration time.Duration
	ctx           context.Context
	limiterID     string
}

func NewRedisLeakyBucket(limiterID string, redis database.RedisStore, rate float64, capacity int, keyExpiration time.Duration) RateLimiter {
	if rate <= 0 || capacity <= 0 || keyExpiration <= 0 || redis == nil {
		log.Fatalf("[FATAL] Invalid parameters for RedisLeakyBucket `%s`", limiterID)
	}
	return &RedisLeakyBucket{
		redis:         redis,
		rate:          rate,
		capacity:      capacity,
		keyExpiration: keyExpiration,
		ctx:           context.Background(),
		limiterID:     limiterID,
	}
}

func (lb *RedisLeakyBucket) Allow(key string) (bool, time.Duration, error) {
	allowed, retryAfter, err := lb.checkBucket(key)
	if err != nil {
		return false, 0, err
	}
	if allowed {
		return true, 0, nil
	}
	return false, time.Duration(retryAfter) * time.Second, nil
}

func (lb *RedisLeakyBucket) checkBucket(key string) (bool, time.Duration, error) {
	script := redis.NewScript(luaLeakyBucket)

	now := float64(time.Now().Unix())
	keys := []string{lb.getRedisKey(key)}
	args := []interface{}{lb.rate, lb.capacity, now, int(lb.keyExpiration.Seconds())}

	result, err := script.Run(lb.ctx, lb.redis.GetNativeInstance().(*redis.Client), keys, args...).Result()
	if err != nil {
		return false, 0, fmt.Errorf("failed to run Lua script: %w", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok {
		return false, 0, fmt.Errorf("unexpected result type or length: %T", result)
	}

	allowed, ok := resultSlice[0].(int64)
	if !ok {
		return false, 0, fmt.Errorf("unexpected 'allowed' type: %T", resultSlice[0])
	}

	var retryAfterSeconds int64
	if allowed != 1 {
		retryAfterSeconds, ok = resultSlice[1].(int64)
		if !ok {
			return false, 0, fmt.Errorf("unexpected 'retry_after' type: %T", resultSlice[1])
		}
	}

	return allowed == 1, time.Duration(retryAfterSeconds), nil
}

func (lb *RedisLeakyBucket) getRedisKey(key string) string {
	return fmt.Sprintf("ratelimit:leaky_bucket:%s:%s", lb.limiterID, key)
}

var luaLeakyBucket = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local expire = tonumber(ARGV[4])

local bucket = redis.call("GET", key)
local tokens = 0
local last_leak = now

if bucket then
    local data = cjson.decode(bucket)
    tokens = data.tokens
    last_leak = data.last_leak
end

local elapsed = now - last_leak
local leaks = elapsed * rate
tokens = math.max(0, tokens - leaks)
last_leak = now

local allowed = 0
local retry_after = 0

if tokens < capacity then
    tokens = tokens + 1
    allowed = 1
else
    retry_after = (tokens - capacity + 1) / rate
end

local new_bucket = cjson.encode({tokens=tokens, last_leak=last_leak})
redis.call("SET", key, new_bucket, "EX", expire)

return {allowed, retry_after}
`
