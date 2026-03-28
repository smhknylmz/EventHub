package worker

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/smhknylmz/EventHub/internal/notification"
)

type RetryPoller struct {
	repo     notification.Repository
	queue    notification.Queue
	logger   *log.Entry
	interval time.Duration
}

func NewRetryPoller(repo notification.Repository, queue notification.Queue, logger *log.Entry, interval time.Duration) *RetryPoller {
	return &RetryPoller{
		repo:     repo,
		queue:    queue,
		logger:   logger.WithField("component", "retrypoller"),
		interval: interval,
	}
}

func (p *RetryPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *RetryPoller) poll(ctx context.Context) {
	notifications, err := p.repo.ListRetryable(ctx, 100)
	if err != nil {
		p.logger.WithError(err).Error("failed to list retryable notifications")
		return
	}

	for _, n := range notifications {
		updated, err := p.repo.UpdateStatus(ctx, n.ID, notification.StatusPending)
		if err != nil {
			p.logger.WithError(err).WithField("notificationId", n.ID).Error("failed to reset status to pending")
			continue
		}

		if err := p.queue.Publish(ctx, updated); err != nil {
			p.logger.WithError(err).WithField("notificationId", n.ID).Error("failed to re-publish notification")
			continue
		}

		p.logger.WithFields(log.Fields{
			"notificationId": n.ID,
			"retryCount":     n.RetryCount,
		}).Info("re-queued notification for retry")
	}
}
