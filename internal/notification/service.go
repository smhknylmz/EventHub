package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Filter struct {
	Status    string
	Channel   string
	BatchID   *uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
}

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	CreateBatch(ctx context.Context, notifications []*Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	List(ctx context.Context, f Filter) ([]*Notification, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Notification, error)
}

type NotificationService struct {
	repo Repository
}

func NewService(repo Repository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	if req.Priority == "" {
		req.Priority = PriorityNormal
	}

	n := &Notification{
		ID:        uuid.Must(uuid.NewV7()),
		Recipient: req.Recipient,
		Channel:   req.Channel,
		Content:   req.Content,
		Priority:  req.Priority,
		Status:    StatusPending,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}

	return toResponse(n), nil
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
			ID:        uuid.Must(uuid.NewV7()),
			BatchID:   &batchID,
			Recipient: r.Recipient,
			Channel:   r.Channel,
			Content:   r.Content,
			Priority:  priority,
			Status:    StatusPending,
		}
	}

	if err := s.repo.CreateBatch(ctx, notifications); err != nil {
		return nil, err
	}

	responses := make([]Response, len(notifications))
	for i, n := range notifications {
		responses[i] = *toResponse(n)
	}

	return &BatchCreateResponse{
		BatchID:       batchID.String(),
		Notifications: responses,
		Total:         len(responses),
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
	return toResponse(n), nil
}

func (s *NotificationService) List(ctx context.Context, params ListParams) (*PagedResponse, error) {
	filter := toFilter(params)
	notifications, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]Response, len(notifications))
	for i, n := range notifications {
		responses[i] = *toResponse(n)
	}

	totalPages := total / filter.PageSize
	if total%filter.PageSize > 0 {
		totalPages++
	}

	return &PagedResponse{
		Data:       responses,
		TotalCount: total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *NotificationService) Cancel(ctx context.Context, id string) (*Response, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	n, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if n.Status != StatusPending {
		return nil, ErrNotCancellable
	}
	n, err = s.repo.UpdateStatus(ctx, uid, StatusCancelled)
	if err != nil {
		return nil, err
	}
	return toResponse(n), nil
}

func toResponse(n *Notification) *Response {
	r := &Response{
		ID:        n.ID.String(),
		Recipient: n.Recipient,
		Channel:   n.Channel,
		Content:   n.Content,
		Priority:  n.Priority,
		Status:    n.Status,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
	if n.BatchID != nil {
		r.BatchID = n.BatchID.String()
	}
	return r
}

func toFilter(params ListParams) Filter {
	f := Filter{
		Status:   params.Status,
		Channel:  params.Channel,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	if params.BatchID != "" {
		if uid, err := uuid.Parse(params.BatchID); err == nil {
			f.BatchID = &uid
		}
	}
	if params.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, params.StartDate); err == nil {
			f.StartDate = &t
		}
	}
	if params.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, params.EndDate); err == nil {
			f.EndDate = &t
		}
	}
	return f
}
