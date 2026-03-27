package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type notificationRepo struct {
	db *sqlx.DB
}

func NewNotificationRepo(db *sqlx.DB) *notificationRepo {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, n *entity.Notification) (*entity.Notification, error) {
	query := `
		INSERT INTO notifications (id, user_id, title, body, is_read, created_at)
		VALUES (:id, :user_id, :title, :body, :is_read, :created_at)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, n)
	if err != nil {
		return nil, fmt.Errorf("notificationRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("notificationRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("notificationRepo.Create: no rows returned")
	}
	if err := rows.StructScan(n); err != nil {
		return nil, fmt.Errorf("notificationRepo.Create scan: %w", err)
	}
	return n, nil
}

func (r *notificationRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID,
	); err != nil {
		return nil, 0, fmt.Errorf("notificationRepo.ListByUser count: %w", err)
	}

	var items []*entity.Notification
	if err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("notificationRepo.ListByUser: %w", err)
	}
	return items, total, nil
}

func (r *notificationRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = true, read_at = now()
		WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("notificationRepo.MarkRead: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("notificationRepo.MarkRead rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *notificationRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM notifications WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return fmt.Errorf("notificationRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("notificationRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}
