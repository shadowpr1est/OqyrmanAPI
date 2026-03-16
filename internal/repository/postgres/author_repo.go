package postgres

import (
	"context"
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
	rows.Next()
	if err := rows.StructScan(author); err != nil {
		return nil, fmt.Errorf("authorRepo.Create scan: %w", err)
	}
	return author, nil
}

func (r *authorRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Author, error) {
	var author entity.Author
	query := `SELECT * FROM authors WHERE id = $1`
	if err := r.db.GetContext(ctx, &author, query, id); err != nil {
		return nil, fmt.Errorf("authorRepo.GetByID: %w", err)
	}
	return &author, nil
}

func (r *authorRepo) List(ctx context.Context, limit, offset int) ([]*entity.Author, int, error) {
	var authors []*entity.Author
	var total int
	query := `SELECT * FROM authors ORDER BY name LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &authors, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("authorRepo.List: %w", err)
	}
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM authors`); err != nil {
		return nil, 0, fmt.Errorf("authorRepo.List count: %w", err)
	}
	return authors, total, nil
}

func (r *authorRepo) Update(ctx context.Context, author *entity.Author) (*entity.Author, error) {
	query := `
		UPDATE authors
		SET name = :name, bio = :bio, birth_date = :birth_date,
		    death_date = :death_date, photo_url = :photo_url
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, author)
	if err != nil {
		return nil, fmt.Errorf("authorRepo.Update: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(author); err != nil {
		return nil, fmt.Errorf("authorRepo.Update scan: %w", err)
	}
	return author, nil
}

func (r *authorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM authors WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("authorRepo.Delete: %w", err)
	}
	return nil
}
