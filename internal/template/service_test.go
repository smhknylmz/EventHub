package template_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/smhknylmz/EventHub/internal/template"
	"github.com/smhknylmz/EventHub/internal/template/mock"
)

func newTestService(t *testing.T) (*template.TemplateService, *mock.MockRepository) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	logger := log.NewEntry(log.New())
	svc := template.NewService(repo, logger)
	return svc, repo
}

func TestServiceCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.Create(context.Background(), template.CreateRequest{
			Name: "welcome",
			Body: "Hello {{name}}",
		})

		require.NoError(t, err)
		assert.Equal(t, "welcome", resp.Name)
		assert.Equal(t, "Hello {{name}}", resp.Body)
		assert.NotEmpty(t, resp.ID)
	})

	t.Run("name conflict", func(t *testing.T) {
		svc, repo := newTestService(t)

		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(template.ErrNameConflict)

		_, err := svc.Create(context.Background(), template.CreateRequest{
			Name: "welcome",
			Body: "Hello",
		})

		assert.ErrorIs(t, err, template.ErrNameConflict)
	})
}

func TestServiceGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().GetByID(gomock.Any(), id).Return(&template.Template{ID: id, Name: "welcome", Body: "Hello"}, nil)

		resp, err := svc.GetByID(context.Background(), id)

		require.NoError(t, err)
		assert.Equal(t, id.String(), resp.ID)
		assert.Equal(t, "welcome", resp.Name)
	})

	t.Run("not found", func(t *testing.T) {
		svc, repo := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, template.ErrNotFound)

		_, err := svc.GetByID(context.Background(), id)

		assert.ErrorIs(t, err, template.ErrNotFound)
	})
}

func TestServiceList(t *testing.T) {
	svc, repo := newTestService(t)

	repo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*template.Template{
		{ID: uuid.Must(uuid.NewV7()), Name: "t1", Body: "b1"},
		{ID: uuid.Must(uuid.NewV7()), Name: "t2", Body: "b2"},
	}, 2, nil)

	resp, err := svc.List(context.Background(), template.ListParams{Page: 1, PageSize: 20})

	require.NoError(t, err)
	assert.Equal(t, 2, resp.TotalCount)
	assert.Equal(t, 1, resp.TotalPages)
}

func TestServiceUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&template.Template{ID: id, Name: "updated", Body: "new body"}, nil)

		resp, err := svc.Update(context.Background(), id, template.UpdateRequest{
			Name: "updated",
			Body: "new body",
		})

		require.NoError(t, err)
		assert.Equal(t, "updated", resp.Name)
		assert.Equal(t, "new body", resp.Body)
	})

	t.Run("not found", func(t *testing.T) {
		svc, repo := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, template.ErrNotFound)

		_, err := svc.Update(context.Background(), id, template.UpdateRequest{
			Name: "updated",
			Body: "new body",
		})

		assert.ErrorIs(t, err, template.ErrNotFound)
	})

	t.Run("name conflict", func(t *testing.T) {
		svc, repo := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, template.ErrNameConflict)

		_, err := svc.Update(context.Background(), id, template.UpdateRequest{
			Name: "duplicate",
			Body: "body",
		})

		assert.ErrorIs(t, err, template.ErrNameConflict)
	})
}

func TestServiceDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().Delete(gomock.Any(), id).Return(nil)

		err := svc.Delete(context.Background(), id)

		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		svc, repo := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().Delete(gomock.Any(), id).Return(template.ErrNotFound)

		err := svc.Delete(context.Background(), id)

		assert.ErrorIs(t, err, template.ErrNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		svc, repo := newTestService(t)
		id := uuid.Must(uuid.NewV7())

		repo.EXPECT().Delete(gomock.Any(), id).Return(errors.New("db error"))

		err := svc.Delete(context.Background(), id)

		assert.Error(t, err)
	})
}
