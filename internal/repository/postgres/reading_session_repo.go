package postgres

import (
	"context"
	"database/sql"
	"errors"
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
		INSERT INTO reading_sessions (id, user_id, book_id, current_page, status, updated_at, finished_at)
		VALUES (:id, :user_id, :book_id, :current_page, :status, :updated_at, :finished_at)
		ON CONFLICT (user_id, book_id)
		DO UPDATE SET
			current_page = :current_page,
			status       = :status,
			updated_at   = :updated_at,
			finished_at  = COALESCE(EXCLUDED.finished_at, reading_sessions.finished_at)
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
	err := r.db.GetContext(ctx, &session, `SELECT * FROM reading_sessions WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("readingSessionRepo.GetByID: %w", err)
	}
	return &session, nil
}

func (r *readingSessionRepo) GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSession, error) {
	var session entity.ReadingSession
	err := r.db.GetContext(ctx, &session,
		`SELECT * FROM reading_sessions WHERE user_id = $1 AND book_id = $2`,
		userID, bookID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("readingSessionRepo.GetByUserAndBook: %w", err)
	}
	return &session, nil
}

func (r *readingSessionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSession, error) {
	var sessions []*entity.ReadingSession
	err := r.db.SelectContext(ctx, &sessions,
		`SELECT * FROM reading_sessions WHERE user_id = $1 ORDER BY updated_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("readingSessionRepo.ListByUser: %w", err)
	}
	return sessions, nil
}

func (r *readingSessionRepo) GetByUserAndBookView(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSessionView, error) {
	var v entity.ReadingSessionView
	err := r.db.GetContext(ctx, &v, `
		SELECT rs.id, rs.user_id, rs.book_id,
		       b.title AS book_title, COALESCE(b.cover_url, '') AS book_cover_url,
		       b.total_pages AS book_total_pages,
		       b.author_id, a.name AS author_name,
		       rs.current_page, rs.status, rs.updated_at, rs.finished_at
		FROM reading_sessions rs
		JOIN books   b ON b.id = rs.book_id    AND b.deleted_at IS NULL
		JOIN authors a ON a.id = b.author_id   AND a.deleted_at IS NULL
		WHERE rs.user_id = $1 AND rs.book_id = $2`,
		userID, bookID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("readingSessionRepo.GetByUserAndBookView: %w", err)
	}
	return &v, nil
}

func (r *readingSessionRepo) ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSessionView, error) {
	var items []*entity.ReadingSessionView
	err := r.db.SelectContext(ctx, &items, `
		SELECT rs.id, rs.user_id, rs.book_id,
		       b.title AS book_title, COALESCE(b.cover_url, '') AS book_cover_url,
		       b.total_pages AS book_total_pages,
		       b.author_id, a.name AS author_name,
		       rs.current_page, rs.status, rs.updated_at, rs.finished_at
		FROM reading_sessions rs
		JOIN books   b ON b.id = rs.book_id    AND b.deleted_at IS NULL
		JOIN authors a ON a.id = b.author_id   AND a.deleted_at IS NULL
		WHERE rs.user_id = $1
		ORDER BY rs.updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("readingSessionRepo.ListByUserView: %w", err)
	}
	return items, nil
}

func (r *readingSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM reading_sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("readingSessionRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("readingSessionRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}
