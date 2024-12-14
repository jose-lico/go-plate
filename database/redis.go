package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jose-lico/go-plate/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	GetNativeInstance() interface{}
}

type Redis struct {
	redis *redis.Client
}

func NewRedis(cfg *config.RedisConfig, logger *zap.Logger) (RedisStore, error) {
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
		logger.Fatal("Error parsing Redis URL", zap.String("URL", url), zap.Error(err))
	}

	redis := redis.NewClient(opt)
	ctx := context.Background()

	for attempts := 0; attempts < maxAttempts; attempts++ {
		if err = redis.Ping(ctx).Err(); err != nil {
			if attempts+1 < maxAttempts {
				logger.Warn(fmt.Sprintf("Failed to connect to Redis (attempt %d). Attempting again in %v...", attempts+1, reconnectCooldown), zap.Error(err))
			}
			time.Sleep(reconnectCooldown)
		} else {
			logger.Info("Connected to Redis")
			return &Redis{redis: redis}, nil
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts, error: %w", maxAttempts, err)
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.redis.Set(ctx, key, value, expiration).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.redis.Get(ctx, key).Result()
}

func (r *Redis) SAdd(ctx context.Context, key string, value interface{}) error {
	return r.redis.SAdd(ctx, key, value).Err()
}

func (r *Redis) SRem(ctx context.Context, key string, value interface{}) error {
	return r.redis.SRem(ctx, key, value).Err()
}

func (r *Redis) Del(ctx context.Context, key string) (int64, error) {
	return r.redis.Del(ctx, key).Result()
}

func (r *Redis) GetNativeInstance() interface{} {
	return r.redis
}
