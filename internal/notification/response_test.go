package notification

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToResponse(t *testing.T) {
	t.Run("without batch id", func(t *testing.T) {
		n := &Notification{
			ID:        uuid.Must(uuid.NewV7()),
			Recipient: "test@example.com",
			Channel:   ChannelEmail,
			Content:   "hello",
			Priority:  PriorityNormal,
			Status:    StatusPending,
		}

		r := n.ToResponse()

		assert.Equal(t, n.ID.String(), r.ID)
		assert.Empty(t, r.BatchID)
		assert.Equal(t, n.Recipient, r.Recipient)
	})

	t.Run("with batch id", func(t *testing.T) {
		batchID := uuid.Must(uuid.NewV7())
		n := &Notification{
			ID:        uuid.Must(uuid.NewV7()),
			BatchID:   &batchID,
			Recipient: "test@example.com",
			Channel:   ChannelSMS,
			Content:   "hi",
			Priority:  PriorityHigh,
			Status:    StatusDelivered,
		}

		r := n.ToResponse()

		assert.Equal(t, batchID.String(), r.BatchID)
	})

	t.Run("with next retry at", func(t *testing.T) {
		retryAt := time.Now().Add(time.Minute)
		n := &Notification{
			ID:          uuid.Must(uuid.NewV7()),
			Channel:     ChannelPush,
			Status:      StatusFailed,
			RetryCount:  2,
			NextRetryAt: &retryAt,
		}

		r := n.ToResponse()

		assert.Equal(t, 2, r.RetryCount)
		assert.NotNil(t, r.NextRetryAt)
	})
}
