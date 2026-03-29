package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"

	"github.com/smhknylmz/EventHub/internal/notification"
	notifmock "github.com/smhknylmz/EventHub/internal/notification/mock"
	"github.com/smhknylmz/EventHub/internal/worker"
)

func TestRetryPollerPoll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := notifmock.NewMockRepository(ctrl)
		queue := notifmock.NewMockQueue(ctrl)
		logger := log.NewEntry(log.New())

		id := uuid.Must(uuid.NewV7())
		n := &notification.Notification{ID: id, RetryCount: 1, Channel: notification.ChannelEmail, Priority: notification.PriorityNormal}
		updated := &notification.Notification{ID: id, Status: notification.StatusPending, Channel: notification.ChannelEmail, Priority: notification.PriorityNormal}

		repo.EXPECT().ListRetryable(gomock.Any(), 100).Return([]*notification.Notification{n}, nil)
		repo.EXPECT().UpdateStatus(gomock.Any(), id, notification.StatusPending).Return(updated, nil)
		queue.EXPECT().Publish(gomock.Any(), updated).Return(nil)

		poller := worker.NewRetryPoller(repo, queue, logger, time.Second)
		poller.Poll(context.Background())
	})

	t.Run("list error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := notifmock.NewMockRepository(ctrl)
		queue := notifmock.NewMockQueue(ctrl)
		logger := log.NewEntry(log.New())

		repo.EXPECT().ListRetryable(gomock.Any(), 100).Return(nil, errors.New("db error"))

		poller := worker.NewRetryPoller(repo, queue, logger, time.Second)
		poller.Poll(context.Background())
	})

	t.Run("update status error continues", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := notifmock.NewMockRepository(ctrl)
		queue := notifmock.NewMockQueue(ctrl)
		logger := log.NewEntry(log.New())

		id1 := uuid.Must(uuid.NewV7())
		id2 := uuid.Must(uuid.NewV7())
		n1 := &notification.Notification{ID: id1}
		n2 := &notification.Notification{ID: id2, Channel: notification.ChannelSMS, Priority: notification.PriorityHigh}
		updated2 := &notification.Notification{ID: id2, Status: notification.StatusPending, Channel: notification.ChannelSMS, Priority: notification.PriorityHigh}

		repo.EXPECT().ListRetryable(gomock.Any(), 100).Return([]*notification.Notification{n1, n2}, nil)
		repo.EXPECT().UpdateStatus(gomock.Any(), id1, notification.StatusPending).Return(nil, errors.New("update fail"))
		repo.EXPECT().UpdateStatus(gomock.Any(), id2, notification.StatusPending).Return(updated2, nil)
		queue.EXPECT().Publish(gomock.Any(), updated2).Return(nil)

		poller := worker.NewRetryPoller(repo, queue, logger, time.Second)
		poller.Poll(context.Background())
	})

	t.Run("publish error continues", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := notifmock.NewMockRepository(ctrl)
		queue := notifmock.NewMockQueue(ctrl)
		logger := log.NewEntry(log.New())

		id := uuid.Must(uuid.NewV7())
		n := &notification.Notification{ID: id}
		updated := &notification.Notification{ID: id, Status: notification.StatusPending}

		repo.EXPECT().ListRetryable(gomock.Any(), 100).Return([]*notification.Notification{n}, nil)
		repo.EXPECT().UpdateStatus(gomock.Any(), id, notification.StatusPending).Return(updated, nil)
		queue.EXPECT().Publish(gomock.Any(), updated).Return(errors.New("queue down"))

		poller := worker.NewRetryPoller(repo, queue, logger, time.Second)
		poller.Poll(context.Background())
	})
}
