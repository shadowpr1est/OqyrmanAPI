package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
	if err := r.db.GetContext(ctx, &lb, `SELECT * FROM library_books WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("libraryBookRepo.GetByID: %w", err)
	}
	return &lb, nil
}

func (r *libraryBookRepo) ListByLibrary(ctx context.Context, libraryID uuid.UUID) ([]*entity.LibraryBook, error) {
	var items []*entity.LibraryBook
	if err := r.db.SelectContext(ctx, &items,
		`SELECT * FROM library_books WHERE library_id = $1 ORDER BY id`,
		libraryID,
	); err != nil {
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
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23514": // check_violation: available_copies <= total_copies or non-negative
				return nil, fmt.Errorf("%w: %s", entity.ErrValidation, pqErr.Message)
			}
		}
		return nil, fmt.Errorf("libraryBookRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("libraryBookRepo.Update rows error: %w", err)
		}
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(lb); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.Update scan: %w", err)
	}
	return lb, nil
}

func (r *libraryBookRepo) UpdateCopies(ctx context.Context, id uuid.UUID, totalCopies *int, availableCopies *int) (*entity.LibraryBook, error) {
	var lb entity.LibraryBook
	err := r.db.GetContext(ctx, &lb, `
		UPDATE library_books
		SET total_copies     = COALESCE($2, total_copies),
		    available_copies = COALESCE($3, available_copies)
		WHERE id = $1
		RETURNING *`, id, totalCopies, availableCopies)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23514" {
			return nil, fmt.Errorf("%w: %s", entity.ErrValidation, pqErr.Message)
		}
		return nil, fmt.Errorf("libraryBookRepo.UpdateCopies: %w", err)
	}
	return &lb, nil
}

func (r *libraryBookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM library_books
		WHERE id = $1
		  AND NOT EXISTS (
		      SELECT 1 FROM reservations
		      WHERE library_book_id = $1
		        AND status IN ('active', 'pending')
		  )`, id)
	if err != nil {
		return fmt.Errorf("libraryBookRepo.Delete: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		var exists bool
		if err := r.db.GetContext(ctx, &exists,
			`SELECT EXISTS(SELECT 1 FROM library_books WHERE id = $1)`, id); err != nil {
			return fmt.Errorf("libraryBookRepo.Delete exists check: %w", err)
		}
		if !exists {
			return entity.ErrNotFound
		}
		return entity.ErrActiveReservationsExist
	}
	return nil
}

// libraryBookViewQuery is the base SELECT for all LibraryBookView methods.
const libraryBookViewQuery = `
	SELECT lb.id, lb.library_id, l.name AS library_name, l.address AS library_address,
	       l.lat AS library_lat, l.lng AS library_lng, l.phone AS library_phone,  
	       
	       lb.book_id, b.title AS book_title, b.isbn AS book_isbn,
		   COALESCE(b.cover_url, '') AS book_cover_url, b.description AS book_description,
		   b.description_kk AS book_description_kk,
		   b.language AS book_language,
		   COALESCE(b.year, 0) AS book_year,
		   COALESCE(b.total_pages,0) AS book_total_pages,
		   b.avg_rating AS book_avg_rating,
		   
	       b.author_id, a.name AS author_name, 
	       a.bio AS author_bio,
	       a.bio_kk AS author_bio_kk,
	       a.birth_date AS author_birth_date,
	       a.death_date AS author_death_date,
	       a.photo_url AS author_photo_url,
	       
	       b.genre_id, g.name AS genre_name, g.slug AS genre_slug,
	       
	       
	       lb.total_copies, lb.available_copies
	FROM library_books lb
	JOIN libraries l ON l.id = lb.library_id AND l.deleted_at IS NULL
	JOIN books     b ON b.id = lb.book_id    AND b.deleted_at IS NULL
	JOIN authors   a ON a.id = b.author_id   AND a.deleted_at IS NULL
	JOIN genres    g ON g.id = b.genre_id    AND g.deleted_at IS NULL`

func (r *libraryBookRepo) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.LibraryBookView, error) {
	var v entity.LibraryBookView
	if err := r.db.GetContext(ctx, &v,
		libraryBookViewQuery+" WHERE lb.id = $1", id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("libraryBookRepo.GetByIDView: %w", err)
	}
	return &v, nil
}

func (r *libraryBookRepo) ListByLibraryView(ctx context.Context, libraryID uuid.UUID) ([]*entity.LibraryBookView, error) {
	var items []*entity.LibraryBookView
	if err := r.db.SelectContext(ctx, &items,
		libraryBookViewQuery+" WHERE lb.library_id = $1 ORDER BY b.title", libraryID,
	); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.ListByLibraryView: %w", err)
	}
	return items, nil
}

func (r *libraryBookRepo) ListByBookView(ctx context.Context, bookID uuid.UUID) ([]*entity.LibraryBookView, error) {
	var items []*entity.LibraryBookView
	if err := r.db.SelectContext(ctx, &items,
		libraryBookViewQuery+" WHERE lb.book_id = $1 ORDER BY l.name", bookID,
	); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.ListByBookView: %w", err)
	}
	return items, nil
}

func (r *libraryBookRepo) BookIDsInLibraries(ctx context.Context) (map[uuid.UUID]bool, error) {
	var ids []uuid.UUID
	if err := r.db.SelectContext(ctx, &ids, `
		SELECT DISTINCT lb.book_id
		FROM library_books lb
		JOIN books b ON b.id = lb.book_id AND b.deleted_at IS NULL`,
	); err != nil {
		return nil, fmt.Errorf("libraryBookRepo.BookIDsInLibraries: %w", err)
	}
	m := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m, nil
}

func (r *libraryBookRepo) SearchInLibrary(ctx context.Context, libraryID uuid.UUID, q string, genreID *uuid.UUID, onlyAvailable bool, limit, offset int) ([]*entity.LibraryBookSearchResult, int, error) {
	var genreParam interface{} = nil
	if genreID != nil {
		genreParam = *genreID
	}

	query := `
		SELECT
			lb.id          AS library_book_id,
			b.id           AS book_id,
			b.title,
			b.cover_url,
			b.year,
			a.name         AS author,
			g.name         AS genre,
			lb.total_copies,
			lb.available_copies,
			lb.available_copies > 0 AS is_available
		FROM library_books lb
		JOIN books    b ON b.id = lb.book_id    AND b.deleted_at IS NULL
		JOIN authors  a ON a.id = b.author_id   AND a.deleted_at IS NULL
		JOIN genres   g ON g.id = b.genre_id    AND g.deleted_at IS NULL
		WHERE lb.library_id = $1
		  AND ($2 = '' OR b.title ILIKE '%' || $2 || '%' OR a.name ILIKE '%' || $2 || '%')
		  AND ($3::uuid IS NULL OR b.genre_id = $3::uuid)
		  AND ($4 = false OR lb.available_copies > 0)
		ORDER BY b.title
		LIMIT $5 OFFSET $6`

	var items []*entity.LibraryBookSearchResult
	if err := r.db.SelectContext(ctx, &items, query, libraryID, q, genreParam, onlyAvailable, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("libraryBookRepo.SearchInLibrary: %w", err)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM library_books lb
		JOIN books   b ON b.id = lb.book_id   AND b.deleted_at IS NULL
		JOIN authors a ON a.id = b.author_id  AND a.deleted_at IS NULL
		WHERE lb.library_id = $1
		  AND ($2 = '' OR b.title ILIKE '%' || $2 || '%' OR a.name ILIKE '%' || $2 || '%')
		  AND ($3::uuid IS NULL OR b.genre_id = $3::uuid)
		  AND ($4 = false OR lb.available_copies > 0)`

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, libraryID, q, genreParam, onlyAvailable); err != nil {
		return nil, 0, fmt.Errorf("libraryBookRepo.SearchInLibrary count: %w", err)
	}

	return items, total, nil
}
