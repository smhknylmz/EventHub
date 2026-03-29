package template

import (
	"context"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	Create(ctx context.Context, t *Template) error
	GetByID(ctx context.Context, id uuid.UUID) (*Template, error)
	List(ctx context.Context, params ListParams) ([]*Template, int, error)
	Update(ctx context.Context, t *Template) (*Template, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type TemplateService struct {
	repo   Repository
	logger *log.Entry
}

func NewService(repo Repository, logger *log.Entry) *TemplateService {
	return &TemplateService{
		repo:   repo,
		logger: logger.WithField("component", "template-service"),
	}
}

func (s *TemplateService) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	t := &Template{
		ID:   uuid.Must(uuid.NewV7()),
		Name: req.Name,
		Body: req.Body,
	}

	if err := s.repo.Create(ctx, t); err != nil {
		s.logger.WithFields(log.Fields{"err": err}).Error("failed to create template")
		return nil, err
	}

	return t.ToResponse(), nil
}

func (s *TemplateService) GetByID(ctx context.Context, id uuid.UUID) (*Response, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithFields(log.Fields{"id": id, "err": err}).Error("failed to get template by ID")
		return nil, err
	}
	return t.ToResponse(), nil
}

func (s *TemplateService) List(ctx context.Context, params ListParams) (*PagedResponse, error) {
	params.Parse()
	templates, total, err := s.repo.List(ctx, params)
	if err != nil {
		s.logger.WithFields(log.Fields{"err": err}).Error("failed to list templates")
		return nil, err
	}

	responses := make([]Response, len(templates))
	for i, t := range templates {
		responses[i] = *t.ToResponse()
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

func (s *TemplateService) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Response, error) {
	t, err := s.repo.Update(ctx, &Template{
		ID:   id,
		Name: req.Name,
		Body: req.Body,
	})
	if err != nil {
		s.logger.WithFields(log.Fields{"id": id, "err": err}).Error("failed to update template")
		return nil, err
	}

	return t.ToResponse(), nil
}

func (s *TemplateService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.WithFields(log.Fields{"id": id, "err": err}).Error("failed to delete template")
		return err
	}

	return nil
}
