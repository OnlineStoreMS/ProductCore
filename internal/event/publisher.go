package event

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"productcore/internal/config"

	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	rdb       *redis.Client
	streamKey string
}

func NewPublisher(rdb *redis.Client, cfg *config.RedisConfig) *Publisher {
	if rdb == nil || !cfg.Enabled {
		return nil
	}
	return &Publisher{rdb: rdb, streamKey: cfg.StreamKey}
}

func (p *Publisher) ProductChanged(ctx context.Context, action string, productID uint64, extra map[string]interface{}) {
	if p == nil {
		return
	}
	payload := map[string]interface{}{
		"action":     action,
		"product_id": productID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range extra {
		payload[k] = v
	}
	raw, _ := json.Marshal(payload)
	_, err := p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamKey,
		Values: map[string]interface{}{"payload": string(raw)},
	}).Result()
	if err != nil {
		log.Printf("event publish failed: %v", err)
	}
}
