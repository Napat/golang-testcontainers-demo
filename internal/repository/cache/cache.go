package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
    client *redis.Client
}

func NewCacheRepository(client *redis.Client) *Cache {
    return &Cache{client: client}
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        return err
    }
    return json.Unmarshal(data, dest)
}
