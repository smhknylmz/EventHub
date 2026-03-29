package notification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/smhknylmz/EventHub/internal/notification"
	"github.com/smhknylmz/EventHub/internal/notification/mock"
)

func newTestService(t *testing.T) (*notification.NotificationService, *mock.MockRepository, *mock.MockQueue) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	queue := mock.NewMockQueue(ctrl)
	logger := log.NewEntry(log.New())
	svc := notification.NewService(repo, queue, logger, 5)
	return svc, repo, queue
}

func TestServiceCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo, queue := newTestService(t)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		queue.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Create(context.Background(), notification.CreateRequest{
			Recipient: "test@example.com",
			Channel:   notification.ChannelEmail,
			Content:   "hello",
		})

		require.NoError(t, err)
		assert.Equal(t, notification.ChannelEmail, resp.Channel)
		assert.Equal(t, notification.StatusPending, resp.Status)
		assert.Equal(t, notification.PriorityNormal, resp.Priority)
	})

	t.Run("explicit priority", func(t *testing.T) {
		svc, repo, queue := newTestService(t)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		queue.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Create(context.Background(), notification.CreateRequest{
			Recipient: "test@example.com",
			Channel:   notification.ChannelSMS,
			Content:   "hello",
			Priority:  notification.PriorityHigh,
		})

		require.NoError(t, err)
		assert.Equal(t, notification.PriorityHigh, resp.Priority)
	})

	t.Run("repo error", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error"))

		_, err := svc.Create(context.Background(), notification.CreateRequest{
			Recipient: "test@example.com",
			Channel:   notification.ChannelEmail,
			Content:   "hello",
		})

		assert.Error(t, err)
	})

	t.Run("queue error marks failed", func(t *testing.T) {
		svc, repo, queue := newTestService(t)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		queue.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(errors.New("queue down"))
		repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), notification.StatusFailed).Return(&notification.Notification{}, nil)

		_, err := svc.Create(context.Background(), notification.CreateRequest{
			Recipient: "test@example.com",
			Channel:   notification.ChannelEmail,
			Content:   "hello",
		})

		assert.Error(t, err)
	})
}

func TestServiceCreateBatch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo, queue := newTestService(t)

		repo.EXPECT().CreateBatch(gomock.Any(), gomock.Any()).Return(nil)
		queue.EXPECT().PublishBatch(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.CreateBatch(context.Background(), notification.BatchCreateRequest{
			Notifications: []notification.CreateRequest{
				{Recipient: "a@b.com", Channel: notification.ChannelEmail, Content: "hi"},
				{Recipient: "+123", Channel: notification.ChannelSMS, Content: "yo"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		assert.NotEmpty(t, resp.BatchID)
	})

	t.Run("queue error marks all failed", func(t *testing.T) {
		svc, repo, queue := newTestService(t)

		repo.EXPECT().CreateBatch(gomock.Any(), gomock.Any()).Return(nil)
		queue.EXPECT().PublishBatch(gomock.Any(), gomock.Any()).Return(errors.New("queue down"))
		repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), notification.StatusFailed).Return(&notification.Notification{}, nil).Times(2)

		_, err := svc.CreateBatch(context.Background(), notification.BatchCreateRequest{
			Notifications: []notification.CreateRequest{
				{Recipient: "a@b.com", Channel: notification.ChannelEmail, Content: "hi"},
				{Recipient: "c@d.com", Channel: notification.ChannelEmail, Content: "yo"},
			},
		})

		assert.Error(t, err)
	})
}

func TestServiceGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().GetByID(gomock.Any(), id).Return(&notification.Notification{ID: id, Status: notification.StatusPending, Channel: notification.ChannelEmail}, nil)

		resp, err := svc.GetByID(context.Background(), id.String())

		require.NoError(t, err)
		assert.Equal(t, id.String(), resp.ID)
	})

	t.Run("invalid id", func(t *testing.T) {
		svc, _, _ := newTestService(t)

		_, err := svc.GetByID(context.Background(), "not-a-uuid")

		assert.ErrorIs(t, err, notification.ErrInvalidID)
	})

	t.Run("not found", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, notification.ErrNotFound)

		_, err := svc.GetByID(context.Background(), id.String())

		assert.ErrorIs(t, err, notification.ErrNotFound)
	})
}

func TestServiceList(t *testing.T) {
	svc, repo, _ := newTestService(t)

	repo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*notification.Notification{
		{ID: uuid.Must(uuid.NewV7()), Status: notification.StatusPending},
		{ID: uuid.Must(uuid.NewV7()), Status: notification.StatusDelivered},
	}, 2, nil)

	resp, err := svc.List(context.Background(), notification.ListParams{Page: 1, PageSize: 20})

	require.NoError(t, err)
	assert.Equal(t, 2, resp.TotalCount)
	assert.Equal(t, 1, resp.TotalPages)
}

func TestServiceCancel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().CancelIfPending(gomock.Any(), id).Return(&notification.Notification{ID: id, Status: notification.StatusCancelled}, nil)

		resp, err := svc.Cancel(context.Background(), id.String())

		require.NoError(t, err)
		assert.Equal(t, notification.StatusCancelled, resp.Status)
	})

	t.Run("invalid id", func(t *testing.T) {
		svc, _, _ := newTestService(t)

		_, err := svc.Cancel(context.Background(), "bad")

		assert.ErrorIs(t, err, notification.ErrInvalidID)
	})

	t.Run("not cancellable", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().CancelIfPending(gomock.Any(), id).Return(nil, notification.ErrNotCancellable)

		_, err := svc.Cancel(context.Background(), id.String())

		assert.ErrorIs(t, err, notification.ErrNotCancellable)
	})
}
