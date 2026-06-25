package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"productcore/internal/dto"

	"github.com/redis/go-redis/v9"
)

const productKeyPrefix = "productcore:product:"

type ProductCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewProductCache(rdb *redis.Client, ttlSeconds int) *ProductCache {
	if rdb == nil {
		return nil
	}
	return &ProductCache{rdb: rdb, ttl: time.Duration(ttlSeconds) * time.Second}
}

func (c *ProductCache) Get(ctx context.Context, id uint64) (*dto.ProductDTO, error) {
	if c == nil {
		return nil, redis.Nil
	}
	data, err := c.rdb.Get(ctx, productKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var item dto.ProductDTO
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (c *ProductCache) Set(ctx context.Context, item *dto.ProductDTO) error {
	if c == nil || item == nil {
		return nil
	}
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, productKey(item.ID), data, c.ttl).Err()
}

func (c *ProductCache) Delete(ctx context.Context, id uint64) error {
	if c == nil {
		return nil
	}
	return c.rdb.Del(ctx, productKey(id)).Err()
}

func productKey(id uint64) string {
	return fmt.Sprintf("%s%d", productKeyPrefix, id)
}
