package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type bookMachineBookRepo struct {
	db *sqlx.DB
}

func NewBookMachineBookRepo(db *sqlx.DB) *bookMachineBookRepo {
	return &bookMachineBookRepo{db: db}
}

func (r *bookMachineBookRepo) Create(ctx context.Context, mb *entity.BookMachineBook) (*entity.BookMachineBook, error) {
	query := `
		INSERT INTO book_machine_books (id, machine_id, book_id, total_copies, available_copies)
		VALUES (:id, :machine_id, :book_id, :total_copies, :available_copies)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, mb)
	if err != nil {
		return nil, fmt.Errorf("bookMachineBookRepo.Create: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(mb); err != nil {
		return nil, fmt.Errorf("bookMachineBookRepo.Create scan: %w", err)
	}
	return mb, nil
}

func (r *bookMachineBookRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.BookMachineBook, error) {
	var mb entity.BookMachineBook
	query := `SELECT * FROM book_machine_books WHERE id = $1`
	if err := r.db.GetContext(ctx, &mb, query, id); err != nil {
		return nil, fmt.Errorf("bookMachineBookRepo.GetByID: %w", err)
	}
	return &mb, nil
}

func (r *bookMachineBookRepo) ListByMachine(ctx context.Context, machineID uuid.UUID) ([]*entity.BookMachineBook, error) {
	var items []*entity.BookMachineBook
	query := `SELECT * FROM book_machine_books WHERE machine_id = $1`
	if err := r.db.SelectContext(ctx, &items, query, machineID); err != nil {
		return nil, fmt.Errorf("bookMachineBookRepo.ListByMachine: %w", err)
	}
	return items, nil
}

func (r *bookMachineBookRepo) ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookMachineBook, error) {
	var items []*entity.BookMachineBook
	query := `SELECT * FROM book_machine_books WHERE book_id = $1`
	if err := r.db.SelectContext(ctx, &items, query, bookID); err != nil {
		return nil, fmt.Errorf("bookMachineBookRepo.ListByBook: %w", err)
	}
	return items, nil
}

func (r *bookMachineBookRepo) Update(ctx context.Context, mb *entity.BookMachineBook) (*entity.BookMachineBook, error) {
	query := `
		UPDATE book_machine_books
		SET total_copies = :total_copies, available_copies = :available_copies
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, mb)
	if err != nil {
		return nil, fmt.Errorf("bookMachineBookRepo.Update: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(mb); err != nil {
		return nil, fmt.Errorf("bookMachineBookRepo.Update scan: %w", err)
	}
	return mb, nil
}

func (r *bookMachineBookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM book_machine_books WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("bookMachineBookRepo.Delete: %w", err)
	}
	return nil
}
