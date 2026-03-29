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
	"github.com/smhknylmz/EventHub/internal/template"
	tmplmock "github.com/smhknylmz/EventHub/internal/template/mock"
)

func newTestService(t *testing.T) (*notification.NotificationService, *mock.MockRepository, *mock.MockQueue) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	queue := mock.NewMockQueue(ctrl)
	logger := log.NewEntry(log.New())
	svc := notification.NewService(repo, queue, nil, logger, 5)
	return svc, repo, queue
}

func newTestServiceWithTemplateRepo(t *testing.T) (*notification.NotificationService, *mock.MockRepository, *mock.MockQueue, *tmplmock.MockRepository) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	queue := mock.NewMockQueue(ctrl)
	tmplRepo := tmplmock.NewMockRepository(ctrl)
	logger := log.NewEntry(log.New())
	svc := notification.NewService(repo, queue, tmplRepo, logger, 5)
	return svc, repo, queue, tmplRepo
}

func TestServiceCreateWithTemplate(t *testing.T) {
	t.Run("success with template and vars", func(t *testing.T) {
		svc, repo, queue, tmplRepo := newTestServiceWithTemplateRepo(t)

		templateID := uuid.Must(uuid.NewV7())
		tmplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(&template.Template{Body: "Hello {{name}}, welcome to {{place}}"}, nil)
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
		queue.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Create(context.Background(), notification.CreateRequest{
			Recipient:    "test@example.com",
			Channel:      notification.ChannelEmail,
			TemplateID:   &templateID,
			TemplateVars: map[string]string{"name": "Semih", "place": "EventHub"},
		})

		require.NoError(t, err)
		assert.Equal(t, "Hello Semih, welcome to EventHub", resp.Content)
	})

	t.Run("template not found", func(t *testing.T) {
		svc, _, _, tmplRepo := newTestServiceWithTemplateRepo(t)

		templateID := uuid.Must(uuid.NewV7())
		tmplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(nil, errors.New("template not found"))

		_, err := svc.Create(context.Background(), notification.CreateRequest{
			Recipient:  "test@example.com",
			Channel:    notification.ChannelEmail,
			TemplateID: &templateID,
		})

		assert.Error(t, err)
	})

	t.Run("unresolved placeholder returns error", func(t *testing.T) {
		svc, _, _, tmplRepo := newTestServiceWithTemplateRepo(t)

		templateID := uuid.Must(uuid.NewV7())
		tmplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(&template.Template{Body: "Hello {{name}}, your code is {{code}}"}, nil)

		_, err := svc.Create(context.Background(), notification.CreateRequest{
			Recipient:    "test@example.com",
			Channel:      notification.ChannelEmail,
			TemplateID:   &templateID,
			TemplateVars: map[string]string{"name": "Semih"},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unresolved template variables")
		assert.Contains(t, err.Error(), "{{code}}")
	})
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
		repo.EXPECT().UpdateStatusBatch(gomock.Any(), gomock.Any(), notification.StatusFailed).Return(nil)

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

		resp, err := svc.GetByID(context.Background(), id)

		require.NoError(t, err)
		assert.Equal(t, id.String(), resp.ID)
	})

	t.Run("not found", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, notification.ErrNotFound)

		_, err := svc.GetByID(context.Background(), id)

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

		resp, err := svc.Cancel(context.Background(), id)

		require.NoError(t, err)
		assert.Equal(t, notification.StatusCancelled, resp.Status)
	})

	t.Run("not cancellable", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().CancelIfPending(gomock.Any(), id).Return(nil, notification.ErrNotCancellable)

		_, err := svc.Cancel(context.Background(), id)

		assert.ErrorIs(t, err, notification.ErrNotCancellable)
	})
}
