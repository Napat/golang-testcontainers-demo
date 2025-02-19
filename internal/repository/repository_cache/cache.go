package repository_cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Napat/golang-testcontainers-demo/pkg/metrics"
	"github.com/go-redis/redis/v8"
)

type CacheRepository struct {
	client  *redis.Client
	metrics *metrics.CacheMetrics
}

func NewCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{
		client:  client,
		metrics: metrics.NewCacheMetrics(),
	}
}

func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	timer := time.Now()
	defer func() {
		r.metrics.OperationDuration.WithLabelValues("set").Observe(time.Since(timer).Seconds())
	}()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *CacheRepository) Get(ctx context.Context, key string, result interface{}) error {
	timer := time.Now()
	defer func() {
		r.metrics.OperationDuration.WithLabelValues("get").Observe(time.Since(timer).Seconds())
	}()

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			r.metrics.MissesTotal.Inc()
		}
		return err
	}

	r.metrics.HitsTotal.Inc()
	return json.Unmarshal(data, result)
}
