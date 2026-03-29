package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	CreateBatch(ctx context.Context, notifications []*Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	List(ctx context.Context, params ListParams) ([]*Notification, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Notification, error)
	CancelIfPending(ctx context.Context, id uuid.UUID) (*Notification, error)
	IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) (*Notification, error)
	ListRetryable(ctx context.Context, limit int) ([]*Notification, error)
}

type Queue interface {
	Publish(ctx context.Context, n *Notification) error
	PublishBatch(ctx context.Context, notifications []*Notification) error
}

type NotificationService struct {
	repo       Repository
	queue      Queue
	logger     *log.Entry
	maxRetries int
}

func NewService(repo Repository, queue Queue, logger *log.Entry, maxRetries int) *NotificationService {
	return &NotificationService{
		repo:       repo,
		queue:      queue,
		logger:     logger.WithField("component", "service"),
		maxRetries: maxRetries,
	}
}

func (s *NotificationService) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	if req.Priority == "" {
		req.Priority = PriorityNormal
	}

	n := &Notification{
		ID:         uuid.Must(uuid.NewV7()),
		Recipient:  req.Recipient,
		Channel:    req.Channel,
		Content:    req.Content,
		Priority:   req.Priority,
		Status:     StatusPending,
		MaxRetries: s.maxRetries,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}

	if err := s.queue.Publish(ctx, n); err != nil {
		s.logger.WithError(err).WithField("notificationId", n.ID).Error("failed to publish to queue")
		if _, updateErr := s.repo.UpdateStatus(ctx, n.ID, StatusFailed); updateErr != nil {
			s.logger.WithError(updateErr).WithField("notificationId", n.ID).Error("failed to update status to failed")
		}
		return nil, err
	}

	return n.ToResponse(), nil
}

func (s *NotificationService) CreateBatch(ctx context.Context, req BatchCreateRequest) (*BatchCreateResponse, error) {
	batchID := uuid.Must(uuid.NewV7())
	notifications := make([]*Notification, len(req.Notifications))

	for i, r := range req.Notifications {
		priority := r.Priority
		if priority == "" {
			priority = PriorityNormal
		}
		notifications[i] = &Notification{
			ID:         uuid.Must(uuid.NewV7()),
			BatchID:    &batchID,
			Recipient:  r.Recipient,
			Channel:    r.Channel,
			Content:    r.Content,
			Priority:   priority,
			Status:     StatusPending,
			MaxRetries: s.maxRetries,
		}
	}

	if err := s.repo.CreateBatch(ctx, notifications); err != nil {
		return nil, err
	}

	if err := s.queue.PublishBatch(ctx, notifications); err != nil {
		s.logger.WithError(err).WithField("batchId", batchID).Error("failed to publish batch to queue")
		for _, n := range notifications {
			if _, updateErr := s.repo.UpdateStatus(ctx, n.ID, StatusFailed); updateErr != nil {
				s.logger.WithError(updateErr).WithField("notificationId", n.ID).Error("failed to update status to failed")
			}
		}
		return nil, err
	}

	return &BatchCreateResponse{
		BatchID: batchID.String(),
		Total:   len(notifications),
	}, nil
}

func (s *NotificationService) GetByID(ctx context.Context, id string) (*Response, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	n, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return n.ToResponse(), nil
}

func (s *NotificationService) List(ctx context.Context, params ListParams) (*PagedResponse, error) {
	params.Parse()
	notifications, total, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	responses := make([]Response, len(notifications))
	for i, n := range notifications {
		responses[i] = *n.ToResponse()
	}

	totalPages := total / params.PageSize
	if total%params.PageSize > 0 {
		totalPages++
	}

	return &PagedResponse{
		Data:       responses,
		TotalCount: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *NotificationService) Cancel(ctx context.Context, id string) (*Response, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	n, err := s.repo.CancelIfPending(ctx, uid)
	if err != nil {
		return nil, err
	}
	return n.ToResponse(), nil
}

