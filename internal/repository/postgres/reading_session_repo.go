package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type readingSessionRepo struct {
	db *sqlx.DB
}

func NewReadingSessionRepo(db *sqlx.DB) *readingSessionRepo {
	return &readingSessionRepo{db: db}
}

func (r *readingSessionRepo) Upsert(ctx context.Context, session *entity.ReadingSession) (*entity.ReadingSession, error) {
	query := `
		INSERT INTO reading_sessions (id, user_id, book_id, current_page, status, updated_at)
		VALUES (:id, :user_id, :book_id, :current_page, :status, :updated_at)
		ON CONFLICT (user_id, book_id)
		DO UPDATE SET current_page = :current_page, status = :status, updated_at = :updated_at
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, session)
	if err != nil {
		return nil, fmt.Errorf("readingSessionRepo.Upsert: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("readingSessionRepo.Upsert rows error: %w", err)
		}
		return nil, fmt.Errorf("readingSessionRepo.Upsert: no rows returned")
	}
	if err := rows.StructScan(session); err != nil {
		return nil, fmt.Errorf("readingSessionRepo.Upsert scan: %w", err)
	}
	return session, nil
}

func (r *readingSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ReadingSession, error) {
	var session entity.ReadingSession
	query := `SELECT * FROM reading_sessions WHERE id = $1`
	if err := r.db.GetContext(ctx, &session, query, id); err != nil {
		return nil, fmt.Errorf("readingSessionRepo.GetByID: %w", err)
	}
	return &session, nil
}

func (r *readingSessionRepo) GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSession, error) {
	var session entity.ReadingSession
	query := `SELECT * FROM reading_sessions WHERE user_id = $1 AND book_id = $2`
	if err := r.db.GetContext(ctx, &session, query, userID, bookID); err != nil {
		return nil, fmt.Errorf("readingSessionRepo.GetByUserAndBook: %w", err)
	}
	return &session, nil
}

func (r *readingSessionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSession, error) {
	var sessions []*entity.ReadingSession
	query := `SELECT * FROM reading_sessions WHERE user_id = $1 ORDER BY updated_at DESC`
	if err := r.db.SelectContext(ctx, &sessions, query, userID); err != nil {
		return nil, fmt.Errorf("readingSessionRepo.ListByUser: %w", err)
	}
	return sessions, nil
}

func (r *readingSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reading_sessions WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("readingSessionRepo.Delete: %w", err)
	}
	return nil
}
