package data

import (
	"companies-service/pkg/config"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func RedisPool(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.GetAddr(),
		Password: cfg.REDIS_PASSWORD,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return rdb, nil
}
