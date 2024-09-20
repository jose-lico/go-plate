package ratelimiting

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jose-lico/go-plate/internal/database"

	"github.com/redis/go-redis/v9"
)

type RedisTokenBucket struct {
	redis         database.RedisStore
	rate          float64
	capacity      float64
	keyExpiration time.Duration
	ctx           context.Context
	limiterID     string
}

func NewRedisTokenBucket(limiterID string, redis database.RedisStore, rate, capacity float64, keyExpiration time.Duration) RateLimiter {
	if rate <= 0 || capacity <= 0 || keyExpiration <= 0 || redis == nil {
		log.Fatalf("[FATAL] Invalid parameters for RedisTokenBucket `%s`", limiterID)
	}

	return &RedisTokenBucket{
		redis:         redis,
		rate:          rate,
		capacity:      capacity,
		keyExpiration: keyExpiration,
		ctx:           context.Background(),
		limiterID:     limiterID,
	}
}

func (tb *RedisTokenBucket) Allow(key string) (bool, time.Duration, error) {
	allowed, retryAfter, err := tb.checkBucket(key)
	if err != nil {
		return false, 0, err
	}

	if allowed {
		return true, 0, nil
	} else {
		return false, time.Duration(retryAfter) * time.Second, nil
	}
}

func (tb *RedisTokenBucket) checkBucket(key string) (bool, time.Duration, error) {
	script := redis.NewScript(luaTokenBucket)

	now := float64(time.Now().UnixNano()) / 1e9
	keys := []string{tb.getRedisKey(key)}
	args := []interface{}{tb.rate, tb.capacity, now, int(tb.keyExpiration.Seconds())}

	result, err := script.Run(tb.ctx, tb.redis.GetNativeInstance().(*redis.Client), keys, args...).Result()
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

func (tb *RedisTokenBucket) getRedisKey(key string) string {
	return fmt.Sprintf("ratelimit:token_bucket:%s:%s", tb.limiterID, key)
}

var luaTokenBucket = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local expire = tonumber(ARGV[4])

local bucket = redis.call("GET", key)
local tokens = capacity
local last_refill = now

if bucket then
    local data = cjson.decode(bucket)
    tokens = data.tokens
    last_refill = data.last_refill

	local elapsed = now - last_refill
    tokens = math.min(capacity, tokens + elapsed * rate)
    last_refill = now
end

local allowed = 0
local retry_after = 0

if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
else
    allowed = 0
    retry_after = (1 - tokens) / rate
end

local new_bucket = cjson.encode({tokens=tokens, last_refill=last_refill})
redis.call("SET", key, new_bucket, "EX", expire)

return {allowed, retry_after}
`
