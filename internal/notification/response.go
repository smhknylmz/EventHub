package notification

import "time"

type Response struct {
	ID        string    `json:"id"`
	BatchID   string    `json:"batchId,omitempty"`
	Recipient string    `json:"recipient"`
	Channel   string    `json:"channel"`
	Content   string    `json:"content"`
	Priority  string    `json:"priority"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BatchCreateResponse struct {
	BatchID       string     `json:"batchId"`
	Notifications []Response `json:"notifications"`
	Total         int        `json:"total"`
}

type PagedResponse struct {
	Data       any   `json:"data"`
	TotalCount int   `json:"totalCount"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
