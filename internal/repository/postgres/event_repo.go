package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type eventRepo struct {
	db *sqlx.DB
}

func NewEventRepo(db *sqlx.DB) *eventRepo {
	return &eventRepo{db: db}
}

func (r *eventRepo) List(ctx context.Context, limit, offset int) ([]*entity.Event, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM events WHERE deleted_at IS NULL AND ends_at > now()`,
	); err != nil {
		return nil, 0, fmt.Errorf("eventRepo.List count: %w", err)
	}

	var items []*entity.Event
	if err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM events
		WHERE deleted_at IS NULL AND ends_at > now()
		ORDER BY starts_at ASC
		LIMIT $1 OFFSET $2`, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("eventRepo.List: %w", err)
	}

	return items, total, nil
}

func (r *eventRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	var e entity.Event
	if err := r.db.GetContext(ctx, &e,
		`SELECT * FROM events WHERE id = $1 AND deleted_at IS NULL`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("eventRepo.GetByID: %w", err)
	}
	return &e, nil
}

func (r *eventRepo) Create(ctx context.Context, e *entity.Event) (*entity.Event, error) {
	rows, err := r.db.NamedQueryContext(ctx, `
		INSERT INTO events (id, title, description, cover_url, location, starts_at, ends_at, created_at)
		VALUES (:id, :title, :description, :cover_url, :location, :starts_at, :ends_at, :created_at)
		RETURNING *`, e,
	)
	if err != nil {
		return nil, fmt.Errorf("eventRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("eventRepo.Create rows: %w", err)
		}
		return nil, fmt.Errorf("eventRepo.Create: no rows returned")
	}
	if err := rows.StructScan(e); err != nil {
		return nil, fmt.Errorf("eventRepo.Create scan: %w", err)
	}
	return e, nil
}

func (r *eventRepo) Update(ctx context.Context, e *entity.Event) (*entity.Event, error) {
	rows, err := r.db.NamedQueryContext(ctx, `
		UPDATE events
		SET title       = :title,
		    description = :description,
		    cover_url   = :cover_url,
		    location    = :location,
		    starts_at   = :starts_at,
		    ends_at     = :ends_at
		WHERE id = :id AND deleted_at IS NULL
		RETURNING *`, e,
	)
	if err != nil {
		return nil, fmt.Errorf("eventRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("eventRepo.Update rows: %w", err)
		}
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(e); err != nil {
		return nil, fmt.Errorf("eventRepo.Update scan: %w", err)
	}
	return e, nil
}

func (r *eventRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE events SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("eventRepo.Delete: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("eventRepo.Delete rows affected: %w", err)
	}
	if n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *eventRepo) FindUpcoming(ctx context.Context, lookahead time.Duration) ([]*entity.Event, error) {
	var items []*entity.Event
	if err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM events
		WHERE deleted_at IS NULL
		  AND reminder_sent = false
		  AND starts_at > now()
		  AND starts_at <= now() + $1::interval
		ORDER BY starts_at ASC`, lookahead.String(),
	); err != nil {
		return nil, fmt.Errorf("eventRepo.FindUpcoming: %w", err)
	}
	return items, nil
}

func (r *eventRepo) MarkReminderSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE events SET reminder_sent = true WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("eventRepo.MarkReminderSent: %w", err)
	}
	return nil
}
