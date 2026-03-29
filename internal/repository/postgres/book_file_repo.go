package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type bookFileRepo struct {
	db *sqlx.DB
}

func NewBookFileRepo(db *sqlx.DB) *bookFileRepo {
	return &bookFileRepo{db: db}
}

func (r *bookFileRepo) Create(ctx context.Context, file *entity.BookFile) (*entity.BookFile, error) {
	query := `
		INSERT INTO book_files (id, book_id, format, file_url, is_audio)
		VALUES (:id, :book_id, :format, :file_url, :is_audio)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, file)
	if err != nil {
		// UNIQUE (book_id, is_audio) violation — file of this type already exists for the book.
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, entity.ErrFileLimitExceeded
		}
		return nil, fmt.Errorf("bookFileRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("bookFileRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("bookFileRepo.Create: no rows returned")
	}
	if err := rows.StructScan(file); err != nil {
		return nil, fmt.Errorf("bookFileRepo.Create scan: %w", err)
	}
	return file, nil
}

func (r *bookFileRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.BookFile, error) {
	var file entity.BookFile
	if err := r.db.GetContext(ctx, &file, `SELECT * FROM book_files WHERE id = $1`, id); err != nil {
		return nil, fmt.Errorf("bookFileRepo.GetByID: %w", err)
	}
	return &file, nil
}

func (r *bookFileRepo) ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookFile, error) {
	var files []*entity.BookFile
	if err := r.db.SelectContext(ctx, &files, `SELECT * FROM book_files WHERE book_id = $1`, bookID); err != nil {
		return nil, fmt.Errorf("bookFileRepo.ListByBook: %w", err)
	}
	return files, nil
}

func (r *bookFileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM book_files WHERE id = $1`, id); err != nil {
		return fmt.Errorf("bookFileRepo.Delete: %w", err)
	}
	return nil
}
