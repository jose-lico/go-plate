package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-plate/internal/config"

	"github.com/redis/go-redis/v9"
)

const (
	reconnectCooldown = 3 * time.Second
	maxAttempts       = 10
)

type RedisStore interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	SAdd(ctx context.Context, key string, value interface{}) error
	SRem(ctx context.Context, key string, value interface{}) error
	Del(ctx context.Context, key string) (int64, error)
}

type Redis struct {
	rdb *redis.Client
}

func NewRedis(cfg *config.RedisConfig) (RedisStore, error) {
	url := fmt.Sprintf("redis%s://default:%s@%s:%s",
		func() string {
			if cfg.UseTLS {
				return "s"
			}
			return ""
		}(),
		cfg.Password, cfg.Host, cfg.Port)

	opt, err := redis.ParseURL(url)

	if err != nil {
		log.Fatalf("[FATAL] Error parsing Redis URL %s: %v", url, err)
	}

	rdb := redis.NewClient(opt)
	ctx := context.Background()

	for attempts := 0; attempts < maxAttempts; attempts++ {
		if err = rdb.Ping(ctx).Err(); err != nil {
			if attempts+1 < maxAttempts {
				log.Printf("[ERROR] Failed to connect to Redis (attempt %d). Error: %v. Attempting again in %v...", attempts+1, err, reconnectCooldown)
			} else {
				log.Printf("[ERROR] Failed to connect to Redis (attempt %d). Error: %v", attempts+1, err)
			}
			time.Sleep(reconnectCooldown)
		} else {
			log.Println("[TRACE] Connected to Redis")
			return &Redis{rdb: rdb}, nil
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts, error: %w", maxAttempts, err)
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiration).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(ctx, key).Result()
}

func (r *Redis) SAdd(ctx context.Context, key string, value interface{}) error {
	return r.rdb.SAdd(ctx, key, value).Err()
}

func (r *Redis) SRem(ctx context.Context, key string, value interface{}) error {
	return r.rdb.SRem(ctx, key, value).Err()
}

func (r *Redis) Del(ctx context.Context, key string) (int64, error) {
	return r.rdb.Del(ctx, key).Result()
}
