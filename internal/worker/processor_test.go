package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/smhknylmz/EventHub/internal/notification"
	notifmock "github.com/smhknylmz/EventHub/internal/notification/mock"
	"github.com/smhknylmz/EventHub/internal/worker"
	workermock "github.com/smhknylmz/EventHub/internal/worker/mock"
	pkgmetrics "github.com/smhknylmz/EventHub/pkg/metrics"
)

func newTestProcessor(t *testing.T) (*worker.Processor, *notifmock.MockRepository, *workermock.MockRateLimiter, *workermock.MockWebhookSender) {
	pkgmetrics.Setup()
	worker.InitMetrics(nil)
	ctrl := gomock.NewController(t)
	repo := notifmock.NewMockRepository(ctrl)
	rl := workermock.NewMockRateLimiter(ctrl)
	wh := workermock.NewMockWebhookSender(ctrl)
	logger := log.NewEntry(log.New())
	p := worker.NewProcessor(repo, rl, wh, logger, time.Second)
	return p, repo, rl, wh
}

func TestProcessSuccess(t *testing.T) {
	p, repo, rl, wh := newTestProcessor(t)
	id := uuid.Must(uuid.NewV7())

	rl.EXPECT().Allow(gomock.Any(), notification.ChannelEmail).Return(true, nil)
	repo.EXPECT().UpdateStatus(gomock.Any(), id, notification.StatusProcessing).Return(
		&notification.Notification{ID: id, Recipient: "test@example.com", Channel: notification.ChannelEmail, Content: "hi", MaxRetries: 5}, nil,
	)
	wh.EXPECT().Send(gomock.Any(), "test@example.com", notification.ChannelEmail, "hi").Return(nil)
	repo.EXPECT().UpdateStatus(gomock.Any(), id, notification.StatusDelivered).Return(&notification.Notification{ID: id}, nil)

	p.Process(context.Background(), &notification.Notification{ID: id, Channel: notification.ChannelEmail, Priority: notification.PriorityHigh})
}

func TestProcessRateLimiterError(t *testing.T) {
	p, repo, rl, _ := newTestProcessor(t)
	id := uuid.Must(uuid.NewV7())

	rl.EXPECT().Allow(gomock.Any(), notification.ChannelEmail).Return(false, errors.New("redis down"))
	repo.EXPECT().UpdateStatus(gomock.Any(), id, notification.StatusFailed).Return(&notification.Notification{ID: id}, nil)

	p.Process(context.Background(), &notification.Notification{ID: id, Channel: notification.ChannelEmail, Priority: notification.PriorityNormal})
}

func TestProcessWebhookFailRetry(t *testing.T) {
	p, repo, rl, wh := newTestProcessor(t)
	id := uuid.Must(uuid.NewV7())

	rl.EXPECT().Allow(gomock.Any(), notification.ChannelSMS).Return(true, nil)
	repo.EXPECT().UpdateStatus(gomock.Any(), id, notification.StatusProcessing).Return(
		&notification.Notification{ID: id, RetryCount: 0, MaxRetries: 5, Recipient: "r", Channel: notification.ChannelSMS, Content: "c"}, nil,
	)
	wh.EXPECT().Send(gomock.Any(), "r", notification.ChannelSMS, "c").Return(errors.New("webhook error"))
	repo.EXPECT().IncrementRetry(gomock.Any(), id, gomock.Any()).Return(&notification.Notification{ID: id, RetryCount: 1}, nil)

	p.Process(context.Background(), &notification.Notification{ID: id, Channel: notification.ChannelSMS, Priority: notification.PriorityNormal})
}

func TestProcessWebhookFailDeadLetter(t *testing.T) {
	p, repo, rl, wh := newTestProcessor(t)
	id := uuid.Must(uuid.NewV7())

	rl.EXPECT().Allow(gomock.Any(), notification.ChannelPush).Return(true, nil)
	repo.EXPECT().UpdateStatus(gomock.Any(), id, notification.StatusProcessing).Return(
		&notification.Notification{ID: id, RetryCount: 4, MaxRetries: 5, Recipient: "r", Channel: notification.ChannelPush, Content: "c"}, nil,
	)
	wh.EXPECT().Send(gomock.Any(), "r", notification.ChannelPush, "c").Return(errors.New("webhook error"))
	repo.EXPECT().UpdateStatus(gomock.Any(), id, notification.StatusDeadLetter).Return(&notification.Notification{ID: id}, nil)

	p.Process(context.Background(), &notification.Notification{ID: id, Channel: notification.ChannelPush, Priority: notification.PriorityLow})
}

func TestHandleFailureExponentialBackoff(t *testing.T) {
	p, repo, _, _ := newTestProcessor(t)
	id := uuid.Must(uuid.NewV7())

	var capturedRetryAt time.Time
	repo.EXPECT().IncrementRetry(gomock.Any(), id, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ uuid.UUID, retryAt time.Time) (*notification.Notification, error) {
			capturedRetryAt = retryAt
			return &notification.Notification{ID: id}, nil
		},
	)

	before := time.Now()
	logger := log.NewEntry(log.New())
	p.HandleFailure(context.Background(), &notification.Notification{ID: id, RetryCount: 2, MaxRetries: 5}, logger)

	expectedDelay := time.Second * time.Duration(8)
	assert.WithinDuration(t, before.Add(expectedDelay), capturedRetryAt, 100*time.Millisecond)
}
