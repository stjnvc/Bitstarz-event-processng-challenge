package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (float64, bool)
	Set(ctx context.Context, key string, value float64, expiration time.Duration) error
	Ping(ctx context.Context) error
	Close() error
}

type RedisCache struct {
	client *redis.Client
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisCache(cfg RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (float64, bool) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, false
		}
		log.Printf("Error getting from Redis key '%s': %v", key, err)
		return 0, false
	}
	fval, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("Error parsing Redis value for key '%s': %v", key, err)
		return 0, false
	}
	return fval, true
}

func (c *RedisCache) Set(ctx context.Context, key string, value float64, expiration time.Duration) error {
	err := c.client.Set(ctx, key, strconv.FormatFloat(value, 'f', -1, 64), expiration).Err()
	if err != nil {
		log.Printf("Error setting Redis key '%s': %v", key, err)
		return fmt.Errorf("failed to set Redis key '%s': %w", key, err)
	}
	return nil
}

func (c *RedisCache) Ping(ctx context.Context) error {
	err := c.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}
	return nil
}

func (c *RedisCache) Close() error {
	err := c.client.Close()
	if err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}
	return nil
}
