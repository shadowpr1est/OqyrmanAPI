package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
		return nil, fmt.Errorf("bookFileRepo.Create: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(file); err != nil {
		return nil, fmt.Errorf("bookFileRepo.Create scan: %w", err)
	}
	return file, nil
}

func (r *bookFileRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.BookFile, error) {
	var file entity.BookFile
	query := `SELECT * FROM book_files WHERE id = $1`
	if err := r.db.GetContext(ctx, &file, query, id); err != nil {
		return nil, fmt.Errorf("bookFileRepo.GetByID: %w", err)
	}
	return &file, nil
}

func (r *bookFileRepo) ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookFile, error) {
	var files []*entity.BookFile
	query := `SELECT * FROM book_files WHERE book_id = $1`
	if err := r.db.SelectContext(ctx, &files, query, bookID); err != nil {
		return nil, fmt.Errorf("bookFileRepo.ListByBook: %w", err)
	}
	return files, nil
}

func (r *bookFileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM book_files WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("bookFileRepo.Delete: %w", err)
	}
	return nil
}
