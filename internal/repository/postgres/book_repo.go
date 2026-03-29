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
		INSERT INTO books (id, author_id, genre_id, title, isbn, cover_url, description, language, year, avg_rating, total_pages)
		VALUES (:id, :author_id, :genre_id, :title, :isbn, :cover_url, :description, :language, :year, :avg_rating, :total_pages)
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
		    language = :language, year = :year, avg_rating = :avg_rating,
		    total_pages = :total_pages
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

func (r *bookRepo) UpdateTotalPages(ctx context.Context, bookID uuid.UUID, totalPages int) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE books SET total_pages = $1 WHERE id = $2 AND deleted_at IS NULL`,
		totalPages, bookID,
	)
	if err != nil {
		return fmt.Errorf("bookRepo.UpdateTotalPages: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("bookRepo.UpdateTotalPages rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
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

func (r *bookRepo) ListPopular(ctx context.Context, limit, offset int) ([]*entity.Book, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM books WHERE deleted_at IS NULL`,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.ListPopular count: %w", err)
	}

	var books []*entity.Book
	if err := r.db.SelectContext(ctx, &books, `
		SELECT * FROM books
		WHERE deleted_at IS NULL
		ORDER BY avg_rating DESC
		LIMIT $1 OFFSET $2`,
		limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.ListPopular: %w", err)
	}

	return books, total, nil
}

func (r *bookRepo) ListSimilar(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.Book, error) {
	var books []*entity.Book
	if err := r.db.SelectContext(ctx, &books, `
		SELECT b.* FROM books b
		JOIN books src ON src.id = $1
		WHERE b.id != $1 AND b.deleted_at IS NULL
		  AND (b.genre_id = src.genre_id OR b.author_id = src.author_id)
		ORDER BY b.avg_rating DESC
		LIMIT $2`,
		bookID, limit,
	); err != nil {
		return nil, fmt.Errorf("bookRepo.ListSimilar: %w", err)
	}
	return books, nil
}

// bookViewQuery is the base SELECT for all BookView methods.
const bookViewQuery = `
	SELECT b.id, b.author_id, a.name AS author_name,
	       b.genre_id, g.name AS genre_name,
	       b.title, b.isbn, COALESCE(b.cover_url, '') AS cover_url,
	       b.description, b.language, b.year, b.avg_rating, b.total_pages, b.created_at
	FROM books b
	JOIN authors a ON a.id = b.author_id AND a.deleted_at IS NULL
	JOIN genres  g ON g.id = b.genre_id  AND g.deleted_at IS NULL
	WHERE b.deleted_at IS NULL`

func (r *bookRepo) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.BookView, error) {
	var v entity.BookView
	err := r.db.GetContext(ctx, &v,
		bookViewQuery+` AND b.id = $1`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("bookRepo.GetByIDView: %w", err)
	}
	return &v, nil
}

func (r *bookRepo) ListView(ctx context.Context, limit, offset int) ([]*entity.BookView, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM books WHERE deleted_at IS NULL`,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.ListView count: %w", err)
	}
	var items []*entity.BookView
	if err := r.db.SelectContext(ctx, &items,
		bookViewQuery+` ORDER BY b.title LIMIT $1 OFFSET $2`,
		limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.ListView: %w", err)
	}
	return items, total, nil
}

func (r *bookRepo) ListByAuthorView(ctx context.Context, authorID uuid.UUID) ([]*entity.BookView, error) {
	var items []*entity.BookView
	if err := r.db.SelectContext(ctx, &items,
		bookViewQuery+` AND b.author_id = $1 ORDER BY b.year DESC`,
		authorID,
	); err != nil {
		return nil, fmt.Errorf("bookRepo.ListByAuthorView: %w", err)
	}
	return items, nil
}

func (r *bookRepo) ListByGenreView(ctx context.Context, genreID uuid.UUID) ([]*entity.BookView, error) {
	var items []*entity.BookView
	if err := r.db.SelectContext(ctx, &items,
		bookViewQuery+` AND b.genre_id = $1 ORDER BY b.title`,
		genreID,
	); err != nil {
		return nil, fmt.Errorf("bookRepo.ListByGenreView: %w", err)
	}
	return items, nil
}

func (r *bookRepo) SearchView(ctx context.Context, query string, limit, offset int) ([]*entity.BookView, int, error) {
	pattern := "%" + query + "%"
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM books b
		JOIN authors a ON a.id = b.author_id AND a.deleted_at IS NULL
		WHERE b.deleted_at IS NULL
		  AND (b.title ILIKE $1 OR b.description ILIKE $1)`,
		pattern,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.SearchView count: %w", err)
	}
	var items []*entity.BookView
	if err := r.db.SelectContext(ctx, &items,
		bookViewQuery+` AND (b.title ILIKE $1 OR b.description ILIKE $1) ORDER BY b.title LIMIT $2 OFFSET $3`,
		pattern, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.SearchView: %w", err)
	}
	return items, total, nil
}

func (r *bookRepo) ListPopularView(ctx context.Context, limit, offset int) ([]*entity.BookView, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM books WHERE deleted_at IS NULL`,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.ListPopularView count: %w", err)
	}
	var items []*entity.BookView
	if err := r.db.SelectContext(ctx, &items,
		bookViewQuery+` ORDER BY b.avg_rating DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("bookRepo.ListPopularView: %w", err)
	}
	return items, total, nil
}

func (r *bookRepo) ListSimilarView(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.BookView, error) {
	var items []*entity.BookView
	if err := r.db.SelectContext(ctx, &items, `
		SELECT b.id, b.author_id, a.name AS author_name,
		       b.genre_id, g.name AS genre_name,
		       b.title, b.isbn, COALESCE(b.cover_url, '') AS cover_url,
		       b.description, b.language, b.year, b.avg_rating, b.total_pages, b.created_at
		FROM books b
		JOIN books src ON src.id = $1
		JOIN authors a ON a.id = b.author_id AND a.deleted_at IS NULL
		JOIN genres  g ON g.id = b.genre_id  AND g.deleted_at IS NULL
		WHERE b.id != $1 AND b.deleted_at IS NULL
		  AND (b.genre_id = src.genre_id OR b.author_id = src.author_id)
		ORDER BY b.avg_rating DESC
		LIMIT $2`,
		bookID, limit,
	); err != nil {
		return nil, fmt.Errorf("bookRepo.ListSimilarView: %w", err)
	}
	return items, nil
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
