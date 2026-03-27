package notification

type CreateRequest struct {
	Recipient string         `json:"recipient" validate:"required"`
	Channel   string         `json:"channel" validate:"required,oneof=email sms push"`
	Content   string         `json:"content" validate:"required"`
	Priority  string         `json:"priority" validate:"omitempty,oneof=high normal low"`
}

type BatchCreateRequest struct {
	Notifications []CreateRequest `json:"notifications" validate:"required,min=1,max=1000,dive"`
}

type ListParams struct {
	Status    string `query:"status"`
	Channel   string `query:"channel"`
	BatchID   string `query:"batchId"`
	StartDate string `query:"startDate"`
	EndDate   string `query:"endDate"`
	Page      int    `query:"page"`
	PageSize  int    `query:"pageSize"`
}
