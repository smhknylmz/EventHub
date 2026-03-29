package template

import "time"

type Response struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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

func (t *Template) ToResponse() *Response {
	return &Response{
		ID:        t.ID.String(),
		Name:      t.Name,
		Body:      t.Body,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
