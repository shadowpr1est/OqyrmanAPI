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

type bookRepo struct {
	db *sqlx.DB
}

func NewBookRepo(db *sqlx.DB) *bookRepo {
	return &bookRepo{db: db}
}

func (r *bookRepo) Create(ctx context.Context, book *entity.Book) (*entity.Book, error) {
	query := `
		INSERT INTO books (id, author_id, genre_id, title, isbn, cover_url, description, language, year, avg_rating)
		VALUES (:id, :author_id, :genre_id, :title, :isbn, :cover_url, :description, :language, :year, :avg_rating)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, book)
	if err != nil {
		return nil, fmt.Errorf("bookRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("bookRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("bookRepo.Create: no rows returned")
	}
	if err := rows.StructScan(book); err != nil {
		return nil, fmt.Errorf("bookRepo.Create scan: %w", err)
	}
	return book, nil
}

func (r *bookRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error) {
	var book entity.Book
	err := r.db.GetContext(ctx, &book,
		`SELECT * FROM books WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("bookRepo.GetByID: %w", err)
	}
	return &book, nil
}

func (r *bookRepo) List(ctx context.Context, limit, offset int) ([]*entity.Book, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM books WHERE deleted_at IS NULL`,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.List count: %w", err)
	}

	var books []*entity.Book
	if err := r.db.SelectContext(ctx, &books, `
		SELECT * FROM books
		WHERE deleted_at IS NULL
		ORDER BY title
		LIMIT $1 OFFSET $2`,
		limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.List: %w", err)
	}

	return books, total, nil
}

func (r *bookRepo) ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*entity.Book, error) {
	var books []*entity.Book
	if err := r.db.SelectContext(ctx, &books, `
		SELECT * FROM books
		WHERE author_id = $1 AND deleted_at IS NULL
		ORDER BY year DESC`,
		authorID,
	); err != nil {
		return nil, fmt.Errorf("bookRepo.ListByAuthor: %w", err)
	}
	return books, nil
}

func (r *bookRepo) ListByGenre(ctx context.Context, genreID uuid.UUID) ([]*entity.Book, error) {
	var books []*entity.Book
	if err := r.db.SelectContext(ctx, &books, `
		SELECT * FROM books
		WHERE genre_id = $1 AND deleted_at IS NULL
		ORDER BY title`,
		genreID,
	); err != nil {
		return nil, fmt.Errorf("bookRepo.ListByGenre: %w", err)
	}
	return books, nil
}

func (r *bookRepo) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Book, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM books
		WHERE deleted_at IS NULL
		  AND (title ILIKE $1 OR description ILIKE $1)`,
		"%"+query+"%",
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.Search count: %w", err)
	}

	var books []*entity.Book
	if err := r.db.SelectContext(ctx, &books, `
		SELECT * FROM books
		WHERE deleted_at IS NULL
		  AND (title ILIKE $1 OR description ILIKE $1)
		ORDER BY title
		LIMIT $2 OFFSET $3`,
		"%"+query+"%", limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.Search: %w", err)
	}

	return books, total, nil
}

func (r *bookRepo) Update(ctx context.Context, book *entity.Book) (*entity.Book, error) {
	query := `
		UPDATE books
		SET author_id = :author_id, genre_id = :genre_id, title = :title,
		    isbn = :isbn, cover_url = :cover_url, description = :description,
		    language = :language, year = :year, avg_rating = :avg_rating
		WHERE id = :id AND deleted_at IS NULL
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, book)
	if err != nil {
		return nil, fmt.Errorf("bookRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("bookRepo.Update rows error: %w", err)
		}
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(book); err != nil {
		return nil, fmt.Errorf("bookRepo.Update scan: %w", err)
	}
	return book, nil
}

func (r *bookRepo) UpdateCoverURL(ctx context.Context, id uuid.UUID, coverURL string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE books SET cover_url = $1 WHERE id = $2 AND deleted_at IS NULL`,
		coverURL, id,
	)
	if err != nil {
		return fmt.Errorf("bookRepo.UpdateCoverURL: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bookRepo.UpdateCoverURL rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *bookRepo) UpdateRating(ctx context.Context, bookID uuid.UUID) error {
	// deleted_at IS NULL — удалённые отзывы не влияют на рейтинг
	_, err := r.db.ExecContext(ctx, `
		UPDATE books
		SET avg_rating = (
			SELECT COALESCE(AVG(rating), 0)
			FROM reviews
			WHERE book_id = $1 AND deleted_at IS NULL
		)
		WHERE id = $1 AND deleted_at IS NULL`,
		bookID,
	)
	if err != nil {
		return fmt.Errorf("bookRepo.UpdateRating: %w", err)
	}
	return nil
}

func (r *bookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE books SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("bookRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bookRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}
