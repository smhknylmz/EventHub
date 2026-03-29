package notification_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smhknylmz/EventHub/internal/notification"
	"github.com/smhknylmz/EventHub/internal/notification/mock"
)

type mockValidator struct{}

func (mockValidator) Validate(any) error { return nil }

func setupEcho(t *testing.T) (*echo.Echo, *mock.MockService) {
	ctrl := gomock.NewController(t)
	svc := mock.NewMockService(ctrl)
	e := echo.New()
	e.Validator = mockValidator{}
	h := notification.NewHandler(svc)
	h.Register(e)
	return e, svc
}

func TestHandlerCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e, svc := setupEcho(t)

		svc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&notification.Response{
			ID: "test-id", Recipient: "test@example.com", Channel: notification.ChannelEmail, Status: notification.StatusPending,
		}, nil)

		body := `{"recipient":"test@example.com","channel":"email","content":"hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "test-id")
	})

	t.Run("bad json", func(t *testing.T) {
		e, _ := setupEcho(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", strings.NewReader("{invalid"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		e, svc := setupEcho(t)

		svc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, notification.ErrNotFound)

		body := `{"recipient":"test@example.com","channel":"email","content":"hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandlerCreateBatch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e, svc := setupEcho(t)

		svc.EXPECT().CreateBatch(gomock.Any(), gomock.Any()).Return(&notification.BatchCreateResponse{BatchID: "batch-123", Total: 1}, nil)

		body := `{"notifications":[{"recipient":"a@b.com","channel":"email","content":"hi"}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/batch", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "batch-123")
	})

	t.Run("bad json", func(t *testing.T) {
		e, _ := setupEcho(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/batch", strings.NewReader("{bad"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestHandlerGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e, svc := setupEcho(t)
		id := uuid.Must(uuid.NewV7())
		svc.EXPECT().GetByID(gomock.Any(), id.String()).Return(&notification.Response{ID: id.String(), Status: notification.StatusDelivered}, nil)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/notifications/%s", id), nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), id.String())
	})

	t.Run("not found", func(t *testing.T) {
		e, svc := setupEcho(t)
		id := uuid.Must(uuid.NewV7())
		svc.EXPECT().GetByID(gomock.Any(), id.String()).Return(nil, notification.ErrNotFound)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/notifications/%s", id), nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandlerList(t *testing.T) {
	e, svc := setupEcho(t)

	svc.EXPECT().List(gomock.Any(), gomock.Any()).Return(&notification.PagedResponse{
		Data: []notification.Response{}, TotalCount: 0, Page: 1, PageSize: 20, TotalPages: 0,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?status=pending&page=1", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "totalCount")
}

func TestHandlerCancel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e, svc := setupEcho(t)
		id := uuid.Must(uuid.NewV7())
		svc.EXPECT().Cancel(gomock.Any(), id.String()).Return(&notification.Response{ID: id.String(), Status: notification.StatusCancelled}, nil)

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/notifications/%s/cancel", id), nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), notification.StatusCancelled)
	})

	t.Run("not cancellable", func(t *testing.T) {
		e, svc := setupEcho(t)
		id := uuid.Must(uuid.NewV7())
		svc.EXPECT().Cancel(gomock.Any(), id.String()).Return(nil, notification.ErrNotCancellable)

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/notifications/%s/cancel", id), nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestHandleError(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid id", notification.ErrInvalidID, http.StatusBadRequest},
		{"not cancellable", notification.ErrNotCancellable, http.StatusBadRequest},
		{"not found", notification.ErrNotFound, http.StatusNotFound},
		{"internal error", errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = notification.HandleError(c, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
