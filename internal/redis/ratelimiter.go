package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	rateLimitPrefix = "ratelimit"
	maxPerSecond    = int64(100)
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (r *RateLimiter) Allow(ctx context.Context, channel string) (bool, error) {
	key := fmt.Sprintf("%s:%s:%d", rateLimitPrefix, channel, time.Now().Unix())

	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("rate limiter incr failed: %w", err)
	}

	if count == 1 {
		r.client.Expire(ctx, key, 2*time.Second)
	}

	return count <= maxPerSecond, nil
}
