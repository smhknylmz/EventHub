package template

type CreateRequest struct {
	Name string `json:"name" validate:"required,max=255"`
	Body string `json:"body" validate:"required,max=10000"`
}

type UpdateRequest struct {
	Name string `json:"name" validate:"required,max=255"`
	Body string `json:"body" validate:"required,max=10000"`
}

type ListParams struct {
	Page     int `query:"page"`
	PageSize int `query:"pageSize"`
}

func (p *ListParams) Parse() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
}
