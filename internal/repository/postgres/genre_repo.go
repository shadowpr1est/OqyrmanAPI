package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type genreRepo struct {
	db *sqlx.DB
}

func NewGenreRepo(db *sqlx.DB) *genreRepo {
	return &genreRepo{db: db}
}

func (r *genreRepo) Create(ctx context.Context, genre *entity.Genre) (*entity.Genre, error) {
	query := `
		INSERT INTO genres (id, name, slug)
		VALUES (:id, :name, :slug)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, genre)
	if err != nil {
		return nil, fmt.Errorf("genreRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("genreRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("genreRepo.Create: no rows returned")
	}
	if err := rows.StructScan(genre); err != nil {
		return nil, fmt.Errorf("genreRepo.Create scan: %w", err)
	}
	return genre, nil
}

func (r *genreRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error) {
	var genre entity.Genre
	query := `SELECT * FROM genres WHERE id = $1`
	if err := r.db.GetContext(ctx, &genre, query, id); err != nil {
		return nil, fmt.Errorf("genreRepo.GetByID: %w", err)
	}
	return &genre, nil
}

func (r *genreRepo) GetBySlug(ctx context.Context, slug string) (*entity.Genre, error) {
	var genre entity.Genre
	query := `SELECT * FROM genres WHERE slug = $1`
	if err := r.db.GetContext(ctx, &genre, query, slug); err != nil {
		return nil, fmt.Errorf("genreRepo.GetBySlug: %w", err)
	}
	return &genre, nil
}

func (r *genreRepo) List(ctx context.Context) ([]*entity.Genre, error) {
	var genres []*entity.Genre
	query := `SELECT * FROM genres ORDER BY name`
	if err := r.db.SelectContext(ctx, &genres, query); err != nil {
		return nil, fmt.Errorf("genreRepo.List: %w", err)
	}
	return genres, nil
}

func (r *genreRepo) Update(ctx context.Context, genre *entity.Genre) (*entity.Genre, error) {
	query := `
		UPDATE genres
		SET name = :name, slug = :slug
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, genre)
	if err != nil {
		return nil, fmt.Errorf("genreRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("genreRepo.Update rows error: %w", err)
		}
		return nil, fmt.Errorf("genreRepo.Update: no rows returned")
	}
	if err := rows.StructScan(genre); err != nil {
		return nil, fmt.Errorf("genreRepo.Update scan: %w", err)
	}
	return genre, nil
}

func (r *genreRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM genres WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("genreRepo.Delete: %w", err)
	}
	return nil
}
