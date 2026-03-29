package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/smhknylmz/EventHub/internal/notification"
)

const (
	streamPrefix = "notifications"
	groupName    = "workers"
)

type Queue struct {
	client *redis.Client
	logger *log.Entry
}

func NewQueue(client *redis.Client, logger *log.Entry) *Queue {
	return &Queue{client: client, logger: logger.WithField("component", "queue")}
}

func AllStreamKeys() []string {
	return []string{
		streamPrefix + ":" + notification.ChannelEmail + ":" + notification.PriorityHigh,
		streamPrefix + ":" + notification.ChannelEmail + ":" + notification.PriorityNormal,
		streamPrefix + ":" + notification.ChannelEmail + ":" + notification.PriorityLow,
		streamPrefix + ":" + notification.ChannelSMS + ":" + notification.PriorityHigh,
		streamPrefix + ":" + notification.ChannelSMS + ":" + notification.PriorityNormal,
		streamPrefix + ":" + notification.ChannelSMS + ":" + notification.PriorityLow,
		streamPrefix + ":" + notification.ChannelPush + ":" + notification.PriorityHigh,
		streamPrefix + ":" + notification.ChannelPush + ":" + notification.PriorityNormal,
		streamPrefix + ":" + notification.ChannelPush + ":" + notification.PriorityLow,
	}
}

func PriorityFromStream(stream string) string {
	parts := strings.Split(stream, ":")
	return parts[len(parts)-1]
}

func (q *Queue) CreateConsumerGroups(ctx context.Context) error {
	for _, key := range AllStreamKeys() {
		err := q.client.XGroupCreateMkStream(ctx, key, groupName, "0").Err()
		if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
			return fmt.Errorf("failed to create consumer group for %s: %w", key, err)
		}
	}
	return nil
}

func (q *Queue) Publish(ctx context.Context, n *notification.Notification) error {
	key := streamPrefix + ":" + n.Channel + ":" + n.Priority
	return q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: map[string]any{
			"id":        n.ID.String(),
			"recipient": n.Recipient,
			"channel":   n.Channel,
			"content":   n.Content,
			"priority":  n.Priority,
		},
	}).Err()
}

func (q *Queue) PublishBatch(ctx context.Context, notifications []*notification.Notification) error {
	pipe := q.client.Pipeline()
	for _, n := range notifications {
		key := streamPrefix + ":" + n.Channel + ":" + n.Priority
		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: key,
			Values: map[string]any{
				"id":        n.ID.String(),
				"recipient": n.Recipient,
				"channel":   n.Channel,
				"content":   n.Content,
				"priority":  n.Priority,
			},
		})
	}
	_, err := pipe.Exec(ctx)
	return err
}

type StreamMessage struct {
	Notification *notification.Notification
	StreamMsgID  string
}

func (q *Queue) Read(ctx context.Context, stream, consumer string, count int64) ([]StreamMessage, error) {
	results, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    10 * time.Second,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	var messages []StreamMessage
	for _, msg := range results[0].Messages {
		idStr, _ := msg.Values["id"].(string)
		id, err := uuid.Parse(idStr)
		if err != nil {
			q.logger.WithError(err).WithField("messageId", msg.ID).Error("invalid notification id in stream")
			q.client.XAck(ctx, stream, groupName, msg.ID)
			continue
		}

		recipient, _ := msg.Values["recipient"].(string)
		channel, _ := msg.Values["channel"].(string)
		content, _ := msg.Values["content"].(string)
		priority, _ := msg.Values["priority"].(string)

		messages = append(messages, StreamMessage{
			Notification: &notification.Notification{
				ID:        id,
				Recipient: recipient,
				Channel:   channel,
				Content:   content,
				Priority:  priority,
			},
			StreamMsgID: msg.ID,
		})
	}

	return messages, nil
}

func (q *Queue) Ack(ctx context.Context, stream, msgID string) error {
	return q.client.XAck(ctx, stream, groupName, msgID).Err()
}

func (q *Queue) TotalDepth(ctx context.Context) (int64, error) {
	var total int64
	for _, key := range AllStreamKeys() {
		length, err := q.client.XLen(ctx, key).Result()
		if err != nil {
			return 0, err
		}
		total += length
	}
	return total, nil
}
