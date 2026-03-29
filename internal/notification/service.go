package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/smhknylmz/EventHub/internal/template"
)

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	CreateBatch(ctx context.Context, notifications []*Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	List(ctx context.Context, params ListParams) ([]*Notification, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Notification, error)
	UpdateStatusBatch(ctx context.Context, ids []uuid.UUID, status string) error
	CancelIfPending(ctx context.Context, id uuid.UUID) (*Notification, error)
	IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) (*Notification, error)
	ListRetryable(ctx context.Context, limit int) ([]*Notification, error)
}

type Queue interface {
	Publish(ctx context.Context, n *Notification) error
	PublishBatch(ctx context.Context, notifications []*Notification) error
}

type NotificationService struct {
	repo         Repository
	queue        Queue
	templateRepo template.Repository
	logger       *log.Entry
	maxRetries   int
}

func NewService(repo Repository, queue Queue, templateRepo template.Repository, logger *log.Entry, maxRetries int) *NotificationService {
	return &NotificationService{
		repo:         repo,
		queue:        queue,
		templateRepo: templateRepo,
		logger:       logger.WithField("component", "service"),
		maxRetries:   maxRetries,
	}
}

func (s *NotificationService) resolveContent(ctx context.Context, req CreateRequest) (string, error) {
	if req.TemplateID == nil {
		return req.Content, nil
	}

	tmpl, err := s.templateRepo.GetByID(ctx, *req.TemplateID)
	if err != nil {
		s.logger.WithFields(log.Fields{"templateId": req.TemplateID, "err": err}).Error("failed to get template")
		return "", err
	}

	return tmpl.RenderBody(req.TemplateVars)
}

func (s *NotificationService) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	if req.Priority == "" {
		req.Priority = PriorityNormal
	}

	content, err := s.resolveContent(ctx, req)
	if err != nil {
		return nil, err
	}

	n := &Notification{
		ID:         uuid.Must(uuid.NewV7()),
		TemplateID: req.TemplateID,
		Recipient:  req.Recipient,
		Channel:    req.Channel,
		Content:    content,
		Priority:   req.Priority,
		Status:     StatusPending,
		MaxRetries: s.maxRetries,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		s.logger.WithFields(log.Fields{"err": err}).Error("failed to create notification in repo")
		return nil, err
	}

	if err := s.queue.Publish(ctx, n); err != nil {
		s.logger.WithFields(log.Fields{"notificationId": n.ID, "err": err}).Error("failed to publish to queue")
		if _, updateErr := s.repo.UpdateStatus(ctx, n.ID, StatusFailed); updateErr != nil {
			s.logger.WithFields(log.Fields{"notificationId": n.ID, "err": updateErr}).Error("failed to update status to failed")
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
		content, err := s.resolveContent(ctx, r)
		if err != nil {
			return nil, err
		}
		notifications[i] = &Notification{
			ID:         uuid.Must(uuid.NewV7()),
			BatchID:    &batchID,
			TemplateID: r.TemplateID,
			Recipient:  r.Recipient,
			Channel:    r.Channel,
			Content:    content,
			Priority:   priority,
			Status:     StatusPending,
			MaxRetries: s.maxRetries,
		}
	}

	if err := s.repo.CreateBatch(ctx, notifications); err != nil {
		s.logger.WithFields(log.Fields{"err": err}).Error("failed to create batch in repo")
		return nil, err
	}

	if err := s.queue.PublishBatch(ctx, notifications); err != nil {
		s.logger.WithFields(log.Fields{"batchId": batchID, "err": err}).Error("failed to publish batch to queue")
		ids := make([]uuid.UUID, len(notifications))
		for i, n := range notifications {
			ids[i] = n.ID
		}
		if updateErr := s.repo.UpdateStatusBatch(ctx, ids, StatusFailed); updateErr != nil {
			s.logger.WithFields(log.Fields{"batchId": batchID, "err": updateErr}).Error("failed to update batch status to failed")
		}
		return nil, err
	}

	return &BatchCreateResponse{
		BatchID: batchID.String(),
		Total:   len(notifications),
	}, nil
}

func (s *NotificationService) GetByID(ctx context.Context, id uuid.UUID) (*Response, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithFields(log.Fields{"notificationId": id, "err": err}).Error("failed to get notification by id")
		return nil, err
	}
	return n.ToResponse(), nil
}

func (s *NotificationService) List(ctx context.Context, params ListParams) (*PagedResponse, error) {
	params.Parse()
	notifications, total, err := s.repo.List(ctx, params)
	if err != nil {
		s.logger.WithFields(log.Fields{"err": err}).Error("failed to list notifications")
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

func (s *NotificationService) Cancel(ctx context.Context, id uuid.UUID) (*Response, error) {
	n, err := s.repo.CancelIfPending(ctx, id)
	if err != nil {
		s.logger.WithFields(log.Fields{"notificationId": id, "err": err}).Error("failed to cancel notification")
		return nil, err
	}
	return n.ToResponse(), nil
}
