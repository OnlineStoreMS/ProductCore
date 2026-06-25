package cache

import (
	"context"
	"fmt"
	"log"

	"productcore/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	log.Printf("redis connected: %s", cfg.Addr)
	return rdb, nil
}
