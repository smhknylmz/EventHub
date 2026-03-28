package worker

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/smhknylmz/EventHub/internal/notification"
	redisadapter "github.com/smhknylmz/EventHub/internal/redis"
	"github.com/smhknylmz/EventHub/internal/webhook"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Processor struct {
	repo        notification.Repository
	rateLimiter *redisadapter.RateLimiter
	webhook     *webhook.Provider
	logger      *log.Entry
	backoffBase time.Duration
}

func NewProcessor(repo notification.Repository, rateLimiter *redisadapter.RateLimiter, webhook *webhook.Provider, logger *log.Entry, backoffBase time.Duration) *Processor {
	return &Processor{
		repo:        repo,
		rateLimiter: rateLimiter,
		webhook:     webhook,
		logger:      logger.WithField("component", "processor"),
		backoffBase: backoffBase,
	}
}

func (p *Processor) Process(ctx context.Context, n *notification.Notification) {
	start := time.Now()
	attrs := attribute.String("channel", n.Channel)

	logger := p.logger.WithFields(log.Fields{
		"notificationId": n.ID,
		"channel":         n.Channel,
		"priority":        n.Priority,
	})

	backoff := 10 * time.Millisecond
	for {
		allowed, err := p.rateLimiter.Allow(ctx, n.Channel)
		if err != nil {
			logger.WithError(err).Error("rate limiter error")
			if _, updateErr := p.repo.UpdateStatus(ctx, n.ID, notification.StatusFailed); updateErr != nil {
				logger.WithError(updateErr).Error("failed to update status to failed after rate limiter error")
			}
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

	current, err := p.repo.UpdateStatus(ctx, n.ID, notification.StatusProcessing)
	if err != nil {
		logger.WithError(err).Error("failed to update status to processing")
		return
	}

	if err := p.webhook.Send(ctx, current.Recipient, current.Channel, current.Content); err != nil {
		logger.WithError(err).WithField("retryCount", current.RetryCount).Error("webhook delivery failed")
		FailedCounter.Add(ctx, 1, metric.WithAttributes(attrs))
		p.handleFailure(ctx, current, logger)
		return
	}

	if _, err := p.repo.UpdateStatus(ctx, current.ID, notification.StatusDelivered); err != nil {
		logger.WithError(err).Error("failed to update status to delivered")
		return
	}

	DeliveredCounter.Add(ctx, 1, metric.WithAttributes(attrs))
	LatencyHistogram.Record(ctx, float64(time.Since(start).Milliseconds()), metric.WithAttributes(attrs))
	logger.Info("notification delivered")
}

func (p *Processor) handleFailure(ctx context.Context, n *notification.Notification, logger *log.Entry) {
	nextRetry := n.RetryCount + 1

	if nextRetry >= n.MaxRetries {
		logger.WithField("maxRetries", n.MaxRetries).Warn("max retries reached, moving to dead letter")
		if _, err := p.repo.UpdateStatus(ctx, n.ID, notification.StatusDeadLetter); err != nil {
			logger.WithError(err).Error("failed to update status to dead_letter")
		}
		return
	}

	delay := p.backoffBase * time.Duration(1<<uint(nextRetry))
	nextRetryAt := time.Now().Add(delay)

	if _, err := p.repo.IncrementRetry(ctx, n.ID, nextRetryAt); err != nil {
		logger.WithError(err).Error("failed to increment retry")
		return
	}

	logger.WithFields(log.Fields{
		"nextRetry":   nextRetry,
		"nextRetryAt": nextRetryAt,
	}).Info("scheduled retry")
}
