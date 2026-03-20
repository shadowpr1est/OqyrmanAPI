package postgres

import (
	"context"
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
	query := `SELECT * FROM reading_notes WHERE id = $1`
	if err := r.db.GetContext(ctx, &note, query, id); err != nil {
		return nil, fmt.Errorf("readingNoteRepo.GetByID: %w", err)
	}
	return &note, nil
}

func (r *readingNoteRepo) ListByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) ([]*entity.ReadingNote, error) {
	var notes []*entity.ReadingNote
	query := `SELECT * FROM reading_notes WHERE user_id = $1 AND book_id = $2 ORDER BY page ASC`
	if err := r.db.SelectContext(ctx, &notes, query, userID, bookID); err != nil {
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
		return nil, fmt.Errorf("readingNoteRepo.Update: no rows returned")
	}
	if err := rows.StructScan(note); err != nil {
		return nil, fmt.Errorf("readingNoteRepo.Update scan: %w", err)
	}
	return note, nil
}

func (r *readingNoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reading_notes WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("readingNoteRepo.Delete: %w", err)
	}
	return nil
}
