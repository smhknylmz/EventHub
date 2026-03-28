package notification

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("notification not found")
	ErrInvalidID     = errors.New("invalid notification id")
	ErrNotCancellable = errors.New("only pending notifications can be cancelled")
)

const (
	StatusPending   = "pending"
	StatusCancelled = "cancelled"

	PriorityHigh   = "high"
	PriorityNormal = "normal"
	PriorityLow    = "low"
)

type Notification struct {
	ID        uuid.UUID
	BatchID   *uuid.UUID
	Recipient string
	Channel   string
	Content   string
	Priority  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
