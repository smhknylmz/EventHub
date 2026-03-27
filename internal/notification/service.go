package notification

import "context"

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Response, error)
	CreateBatch(ctx context.Context, req BatchCreateRequest) (*BatchCreateResponse, error)
	Get(ctx context.Context, id string) (*Response, error)
	List(ctx context.Context, params ListParams) (*PagedResponse, error)
	Cancel(ctx context.Context, id string) (*Response, error)
}

type NotificationService struct{}

func NewService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	return nil, nil
}

func (s *NotificationService) CreateBatch(ctx context.Context, req BatchCreateRequest) (*BatchCreateResponse, error) {
	return nil, nil
}

func (s *NotificationService) Get(ctx context.Context, id string) (*Response, error) {
	return nil, nil
}

func (s *NotificationService) List(ctx context.Context, params ListParams) (*PagedResponse, error) {
	return nil, nil
}

func (s *NotificationService) Cancel(ctx context.Context, id string) (*Response, error) {
	return nil, nil
}
