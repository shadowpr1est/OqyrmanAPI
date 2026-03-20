package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type libraryBookRepo struct {
	db *sqlx.DB
}

func NewLibraryBookRepo(db *sqlx.DB) *libraryBookRepo {
	return &libraryBookRepo{db: db}
}

func (r *libraryBookRepo) Create(ctx context.Context, lb *entity.LibraryBook) (*entity.LibraryBook, error) {
	query := `
		INSERT INTO library_books (id, library_id, book_id, total_copies, available_copies)
		VALUES (:id, :library_id, :book_id, :total_copies, :available_copies)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, lb)
	if err != nil {
		return nil, fmt.Errorf("libraryBookRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("libraryBookRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("libraryBookRepo.Create: no rows returned")
	}
	if err := rows.StructScan(lb); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.Create scan: %w", err)
	}
	return lb, nil
}

func (r *libraryBookRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.LibraryBook, error) {
	var lb entity.LibraryBook
	query := `SELECT * FROM library_books WHERE id = $1`
	if err := r.db.GetContext(ctx, &lb, query, id); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.GetByID: %w", err)
	}
	return &lb, nil
}

func (r *libraryBookRepo) ListByLibrary(ctx context.Context, libraryID uuid.UUID) ([]*entity.LibraryBook, error) {
	var items []*entity.LibraryBook
	query := `SELECT * FROM library_books WHERE library_id = $1`
	if err := r.db.SelectContext(ctx, &items, query, libraryID); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.ListByLibrary: %w", err)
	}
	return items, nil
}

func (r *libraryBookRepo) ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.LibraryBook, error) {
	var items []*entity.LibraryBook
	query := `SELECT * FROM library_books WHERE book_id = $1`
	if err := r.db.SelectContext(ctx, &items, query, bookID); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.ListByBook: %w", err)
	}
	return items, nil
}

func (r *libraryBookRepo) Update(ctx context.Context, lb *entity.LibraryBook) (*entity.LibraryBook, error) {
	query := `
		UPDATE library_books
		SET total_copies = :total_copies, available_copies = :available_copies
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, lb)
	if err != nil {
		return nil, fmt.Errorf("libraryBookRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("libraryBookRepo.Update rows error: %w", err)
		}
		return nil, fmt.Errorf("libraryBookRepo.Update: no rows returned")
	}
	if err := rows.StructScan(lb); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.Update scan: %w", err)
	}
	return lb, nil
}

func (r *libraryBookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM library_books WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("libraryBookRepo.Delete: %w", err)
	}
	return nil
}
