package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/smhknylmz/EventHub/internal/template"
)

type TemplateRepo struct {
	pool *pgxpool.Pool
}

func NewTemplateRepo(pool *pgxpool.Pool) *TemplateRepo {
	return &TemplateRepo{pool: pool}
}

func (r *TemplateRepo) Create(ctx context.Context, t *template.Template) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO templates (id, name, body)
		 VALUES ($1, $2, $3)
		 RETURNING created_at, updated_at`,
		t.ID, t.Name, t.Body,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return template.ErrNameConflict
		}
		return err
	}
	return nil
}

func (r *TemplateRepo) GetByID(ctx context.Context, id uuid.UUID) (*template.Template, error) {
	var t template.Template
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, body, created_at, updated_at
		 FROM templates WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Body, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, template.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) List(ctx context.Context, f template.ListParams) ([]*template.Template, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM templates").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.PageSize

	rows, err := r.pool.Query(ctx,
		`SELECT id, name, body, created_at, updated_at
		 FROM templates ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		f.PageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var templates []*template.Template
	for rows.Next() {
		var t template.Template
		if err := rows.Scan(&t.ID, &t.Name, &t.Body, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		templates = append(templates, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *TemplateRepo) Update(ctx context.Context, t *template.Template) (*template.Template, error) {
	var updated template.Template
	err := r.pool.QueryRow(ctx,
		`UPDATE templates SET name = $1, body = $2, updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, name, body, created_at, updated_at`,
		t.Name, t.Body, t.ID,
	).Scan(&updated.ID, &updated.Name, &updated.Body, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, template.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, template.ErrNameConflict
		}
		return nil, err
	}
	return &updated, nil
}

func (r *TemplateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM templates WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return template.ErrNotFound
	}
	return nil
}
