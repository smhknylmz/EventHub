package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

const (
	rateLimitPrefix = "ratelimit"
	maxPerSecond    = int64(100)
)

type RateLimiter struct {
	client *redis.Client
	logger *log.Entry
}

func NewRateLimiter(client *redis.Client, logger *log.Entry) *RateLimiter {
	return &RateLimiter{client: client, logger: logger.WithField("component", "ratelimiter")}
}

func (r *RateLimiter) Allow(ctx context.Context, channel string) (bool, error) {
	key := fmt.Sprintf("%s:%s:%d", rateLimitPrefix, channel, time.Now().Unix())

	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("rate limiter incr failed: %w", err)
	}

	if count == 1 {
		if err := r.client.Expire(ctx, key, 2*time.Second).Err(); err != nil {
			r.logger.WithError(err).WithField("key", key).Error("failed to set rate limit key expiry")
		}
	}

	return count <= maxPerSecond, nil
}
