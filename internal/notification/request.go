package notification

import (
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	Recipient    string            `json:"recipient" validate:"required,max=255"`
	Channel      string            `json:"channel" validate:"required,oneof=email sms push"`
	Content      string            `json:"content" validate:"required_without=TemplateID,max=10000"`
	Priority     string            `json:"priority" validate:"omitempty,oneof=high normal low"`
	TemplateID   *uuid.UUID        `json:"templateId" validate:"omitempty"`
	TemplateVars map[string]string `json:"templateVars" validate:"omitempty,dive,max=1000"`
}

type BatchCreateRequest struct {
	Notifications []CreateRequest `json:"notifications" validate:"required,min=1,max=1000,dive"`
}

type ListParams struct {
	Status    string     `query:"status"`
	Channel   string     `query:"channel"`
	BatchID   *uuid.UUID `query:"-"`
	StartDate *time.Time `query:"-"`
	EndDate   *time.Time `query:"-"`
	Page      int        `query:"page"`
	PageSize  int        `query:"pageSize"`

	RawBatchID   string `query:"batchId"`
	RawStartDate string `query:"startDate"`
	RawEndDate   string `query:"endDate"`
}

func (p *ListParams) Parse() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
	if p.RawBatchID != "" {
		if uid, err := uuid.Parse(p.RawBatchID); err == nil {
			p.BatchID = &uid
		}
	}
	if p.RawStartDate != "" {
		if t, err := time.Parse(time.RFC3339, p.RawStartDate); err == nil {
			p.StartDate = &t
		}
	}
	if p.RawEndDate != "" {
		if t, err := time.Parse(time.RFC3339, p.RawEndDate); err == nil {
			p.EndDate = &t
		}
	}
}
