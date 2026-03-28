package worker

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/smhknylmz/EventHub/internal/notification"
	redisadapter "github.com/smhknylmz/EventHub/internal/redis"
	"github.com/smhknylmz/EventHub/internal/webhook"
)

type Processor struct {
	repo        notification.Repository
	rateLimiter *redisadapter.RateLimiter
	webhook     *webhook.Provider
}

func NewProcessor(repo notification.Repository, rateLimiter *redisadapter.RateLimiter, webhook *webhook.Provider) *Processor {
	return &Processor{
		repo:        repo,
		rateLimiter: rateLimiter,
		webhook:     webhook,
	}
}

func (p *Processor) Process(ctx context.Context, n *notification.Notification) {
	logger := log.WithFields(log.Fields{
		"notification_id": n.ID,
		"channel":         n.Channel,
		"priority":        n.Priority,
	})

	backoff := 10 * time.Millisecond
	for {
		allowed, err := p.rateLimiter.Allow(ctx, n.Channel)
		if err != nil {
			logger.WithError(err).Error("rate limiter error")
			return
		}
		if allowed {
			break
		}
		time.Sleep(backoff)
		if backoff < 500*time.Millisecond {
			backoff *= 2
		}
	}

	if _, err := p.repo.UpdateStatus(ctx, n.ID, notification.StatusProcessing); err != nil {
		logger.WithError(err).Error("failed to update status to processing")
		return
	}

	if err := p.webhook.Send(ctx, n.Recipient, n.Channel, n.Content); err != nil {
		logger.WithError(err).Error("webhook delivery failed")
		if _, updateErr := p.repo.UpdateStatus(ctx, n.ID, notification.StatusFailed); updateErr != nil {
			logger.WithError(updateErr).Error("failed to update status to failed")
		}
		return
	}

	if _, err := p.repo.UpdateStatus(ctx, n.ID, notification.StatusDelivered); err != nil {
		logger.WithError(err).Error("failed to update status to delivered")
		return
	}

	logger.Info("notification delivered")
}
