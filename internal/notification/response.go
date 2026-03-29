package notification

import "time"

type Response struct {
	ID          string     `json:"id"`
	BatchID     string     `json:"batchId,omitempty"`
	TemplateID  string     `json:"templateId,omitempty"`
	Recipient   string     `json:"recipient"`
	Channel     string     `json:"channel"`
	Content     string     `json:"content"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	RetryCount  int        `json:"retryCount"`
	NextRetryAt *time.Time `json:"nextRetryAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type BatchCreateResponse struct {
	BatchID string `json:"batchId"`
	Total   int    `json:"total"`
}

type PagedResponse struct {
	Data       any `json:"data"`
	TotalCount int `json:"totalCount"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func (n *Notification) ToResponse() *Response {
	r := &Response{
		ID:          n.ID.String(),
		Recipient:   n.Recipient,
		Channel:     n.Channel,
		Content:     n.Content,
		Priority:    n.Priority,
		Status:      n.Status,
		RetryCount:  n.RetryCount,
		NextRetryAt: n.NextRetryAt,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
	if n.BatchID != nil {
		r.BatchID = n.BatchID.String()
	}
	if n.TemplateID != nil {
		r.TemplateID = n.TemplateID.String()
	}
	return r
}
