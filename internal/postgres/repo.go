package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/smhknylmz/EventHub/internal/notification"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, n *notification.Notification) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO notifications (id, batch_id, recipient, channel, content, priority, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at, updated_at`,
		n.ID, n.BatchID, n.Recipient, n.Channel, n.Content, n.Priority, n.Status,
	).Scan(&n.CreatedAt, &n.UpdatedAt)
}

func (r *Repo) CreateBatch(ctx context.Context, notifications []*notification.Notification) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, n := range notifications {
		err := tx.QueryRow(ctx,
			`INSERT INTO notifications (id, batch_id, recipient, channel, content, priority, status)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING created_at, updated_at`,
			n.ID, n.BatchID, n.Recipient, n.Channel, n.Content, n.Priority, n.Status,
		).Scan(&n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*notification.Notification, error) {
	var n notification.Notification
	err := r.pool.QueryRow(ctx,
		`SELECT id, batch_id, recipient, channel, content, priority, status, created_at, updated_at
		 FROM notifications WHERE id = $1`, id,
	).Scan(&n.ID, &n.BatchID, &n.Recipient, &n.Channel, &n.Content, &n.Priority, &n.Status, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notification.ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}

func (r *Repo) List(ctx context.Context, f notification.Filter) ([]*notification.Notification, int, error) {
	conditions := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	if f.Channel != "" {
		conditions = append(conditions, fmt.Sprintf("channel = $%d", argIdx))
		args = append(args, f.Channel)
		argIdx++
	}
	if f.BatchID != nil {
		conditions = append(conditions, fmt.Sprintf("batch_id = $%d", argIdx))
		args = append(args, *f.BatchID)
		argIdx++
	}
	if f.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *f.StartDate)
		argIdx++
	}
	if f.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *f.EndDate)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int
	err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s", where), args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)

	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(
			`SELECT id, batch_id, recipient, channel, content, priority, status, created_at, updated_at
			 FROM notifications WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
			where, argIdx, argIdx+1,
		), args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		var n notification.Notification
		if err := rows.Scan(&n.ID, &n.BatchID, &n.Recipient, &n.Channel, &n.Content, &n.Priority, &n.Status, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, &n)
	}

	return notifications, total, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*notification.Notification, error) {
	var n notification.Notification
	err := r.pool.QueryRow(ctx,
		`UPDATE notifications SET status = $1, updated_at = NOW() WHERE id = $2
		 RETURNING id, batch_id, recipient, channel, content, priority, status, created_at, updated_at`,
		status, id,
	).Scan(&n.ID, &n.BatchID, &n.Recipient, &n.Channel, &n.Content, &n.Priority, &n.Status, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notification.ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}
