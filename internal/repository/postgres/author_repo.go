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

type authorRepo struct {
	db *sqlx.DB
}

func NewAuthorRepo(db *sqlx.DB) *authorRepo {
	return &authorRepo{db: db}
}

func (r *authorRepo) Create(ctx context.Context, author *entity.Author) (*entity.Author, error) {
	query := `
		INSERT INTO authors (id, name, bio, birth_date, death_date, photo_url)
		VALUES (:id, :name, :bio, :birth_date, :death_date, :photo_url)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, author)
	if err != nil {
		return nil, fmt.Errorf("authorRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("authorRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("authorRepo.Create: no rows returned")
	}
	if err := rows.StructScan(author); err != nil {
		return nil, fmt.Errorf("authorRepo.Create scan: %w", err)
	}
	return author, nil
}

func (r *authorRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Author, error) {
	var author entity.Author
	err := r.db.GetContext(ctx, &author,
		`SELECT * FROM authors WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("authorRepo.GetByID: %w", err)
	}
	return &author, nil
}

func (r *authorRepo) List(ctx context.Context, limit, offset int) ([]*entity.Author, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM authors WHERE deleted_at IS NULL`,
	); err != nil {
		return nil, 0, fmt.Errorf("authorRepo.List count: %w", err)
	}

	var authors []*entity.Author
	if err := r.db.SelectContext(ctx, &authors, `
		SELECT * FROM authors
		WHERE deleted_at IS NULL
		ORDER BY name
		LIMIT $1 OFFSET $2`,
		limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("authorRepo.List: %w", err)
	}

	return authors, total, nil
}

func (r *authorRepo) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Author, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM authors
		WHERE deleted_at IS NULL
		  AND (name ILIKE $1 OR bio ILIKE $1)`,
		"%"+query+"%",
	); err != nil {
		return nil, 0, fmt.Errorf("authorRepo.Search count: %w", err)
	}

	var authors []*entity.Author
	if err := r.db.SelectContext(ctx, &authors, `
		SELECT * FROM authors
		WHERE deleted_at IS NULL
		  AND (name ILIKE $1 OR bio ILIKE $1)
		ORDER BY name
		LIMIT $2 OFFSET $3`,
		"%"+query+"%", limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("authorRepo.Search: %w", err)
	}

	return authors, total, nil
}

func (r *authorRepo) Update(ctx context.Context, author *entity.Author) (*entity.Author, error) {
	query := `
		UPDATE authors
		SET name = :name, bio = :bio, birth_date = :birth_date,
		    death_date = :death_date, photo_url = :photo_url
		WHERE id = :id AND deleted_at IS NULL
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, author)
	if err != nil {
		return nil, fmt.Errorf("authorRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("authorRepo.Update rows error: %w", err)
		}
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(author); err != nil {
		return nil, fmt.Errorf("authorRepo.Update scan: %w", err)
	}
	return author, nil
}

func (r *authorRepo) UpdatePhotoURL(ctx context.Context, id uuid.UUID, url string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE authors SET photo_url = $1 WHERE id = $2 AND deleted_at IS NULL`, url, id,
	)
	if err != nil {
		return fmt.Errorf("authorRepo.UpdatePhotoURL: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *authorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE authors SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("authorRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("authorRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}
