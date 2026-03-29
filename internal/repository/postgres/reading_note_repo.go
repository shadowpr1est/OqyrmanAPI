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

type readingNoteRepo struct {
	db *sqlx.DB
}

func NewReadingNoteRepo(db *sqlx.DB) *readingNoteRepo {
	return &readingNoteRepo{db: db}
}

func (r *readingNoteRepo) Create(ctx context.Context, note *entity.ReadingNote) (*entity.ReadingNote, error) {
	query := `
		INSERT INTO reading_notes (id, user_id, book_id, page, content, created_at)
		VALUES (:id, :user_id, :book_id, :page, :content, :created_at)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, note)
	if err != nil {
		return nil, fmt.Errorf("readingNoteRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("readingNoteRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("readingNoteRepo.Create: no rows returned")
	}
	if err := rows.StructScan(note); err != nil {
		return nil, fmt.Errorf("readingNoteRepo.Create scan: %w", err)
	}
	return note, nil
}

func (r *readingNoteRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ReadingNote, error) {
	var note entity.ReadingNote
	err := r.db.GetContext(ctx, &note, `SELECT * FROM reading_notes WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("readingNoteRepo.GetByID: %w", err)
	}
	return &note, nil
}

func (r *readingNoteRepo) ListByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) ([]*entity.ReadingNote, error) {
	var notes []*entity.ReadingNote
	err := r.db.SelectContext(ctx, &notes,
		`SELECT * FROM reading_notes WHERE user_id = $1 AND book_id = $2 ORDER BY page ASC`,
		userID, bookID,
	)
	if err != nil {
		return nil, fmt.Errorf("readingNoteRepo.ListByUserAndBook: %w", err)
	}
	return notes, nil
}

func (r *readingNoteRepo) Update(ctx context.Context, note *entity.ReadingNote) (*entity.ReadingNote, error) {
	query := `
		UPDATE reading_notes
		SET page = :page, content = :content
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, note)
	if err != nil {
		return nil, fmt.Errorf("readingNoteRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("readingNoteRepo.Update rows error: %w", err)
		}
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(note); err != nil {
		return nil, fmt.Errorf("readingNoteRepo.Update scan: %w", err)
	}
	return note, nil
}

func (r *readingNoteRepo) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReadingNoteView, error) {
	var v entity.ReadingNoteView
	err := r.db.GetContext(ctx, &v, `
		SELECT n.id, n.user_id, n.book_id, b.title AS book_title,
		       n.page, n.content, n.created_at
		FROM reading_notes n
		JOIN books b ON b.id = n.book_id AND b.deleted_at IS NULL
		WHERE n.id = $1`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("readingNoteRepo.GetByIDView: %w", err)
	}
	return &v, nil
}

func (r *readingNoteRepo) ListByUserAndBookView(ctx context.Context, userID, bookID uuid.UUID) ([]*entity.ReadingNoteView, error) {
	var items []*entity.ReadingNoteView
	err := r.db.SelectContext(ctx, &items, `
		SELECT n.id, n.user_id, n.book_id, b.title AS book_title,
		       n.page, n.content, n.created_at
		FROM reading_notes n
		JOIN books b ON b.id = n.book_id AND b.deleted_at IS NULL
		WHERE n.user_id = $1 AND n.book_id = $2
		ORDER BY n.page ASC`,
		userID, bookID,
	)
	if err != nil {
		return nil, fmt.Errorf("readingNoteRepo.ListByUserAndBookView: %w", err)
	}
	return items, nil
}

func (r *readingNoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM reading_notes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("readingNoteRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("readingNoteRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}
