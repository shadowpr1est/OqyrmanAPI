package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type libraryRepo struct {
	db *sqlx.DB
}

func NewLibraryRepo(db *sqlx.DB) *libraryRepo {
	return &libraryRepo{db: db}
}

func (r *libraryRepo) Create(ctx context.Context, library *entity.Library) (*entity.Library, error) {
	query := `
		INSERT INTO libraries (id, name, address, lat, lng, phone)
		VALUES (:id, :name, :address, :lat, :lng, :phone)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, library)
	if err != nil {
		return nil, fmt.Errorf("libraryRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("libraryRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("libraryRepo.Create: no rows returned")
	}
	if err := rows.StructScan(library); err != nil {
		return nil, fmt.Errorf("libraryRepo.Create scan: %w", err)
	}
	return library, nil
}

func (r *libraryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Library, error) {
	var library entity.Library
	query := `SELECT * FROM libraries WHERE id = $1`
	if err := r.db.GetContext(ctx, &library, query, id); err != nil {
		return nil, fmt.Errorf("libraryRepo.GetByID: %w", err)
	}
	return &library, nil
}

func (r *libraryRepo) List(ctx context.Context, limit, offset int) ([]*entity.Library, int, error) {
	var libraries []*entity.Library
	var total int
	query := `SELECT * FROM libraries ORDER BY name LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &libraries, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("libraryRepo.List: %w", err)
	}
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM libraries`); err != nil {
		return nil, 0, fmt.Errorf("libraryRepo.List count: %w", err)
	}
	return libraries, total, nil
}

func (r *libraryRepo) ListNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Library, error) {
	var libraries []*entity.Library
	// LEAST(1.0, ...) защищает acos от NaN при точном совпадении координат
	query := `
		SELECT * FROM libraries
		WHERE (6371 * acos(LEAST(1.0, cos(radians($1)) * cos(radians(lat)) *
		cos(radians(lng) - radians($2)) + sin(radians($1)) * sin(radians(lat))))) < $3
		ORDER BY name`
	if err := r.db.SelectContext(ctx, &libraries, query, lat, lng, radiusKm); err != nil {
		return nil, fmt.Errorf("libraryRepo.ListNearby: %w", err)
	}
	return libraries, nil
}

func (r *libraryRepo) Update(ctx context.Context, library *entity.Library) (*entity.Library, error) {
	query := `
		UPDATE libraries
		SET name = :name, address = :address, lat = :lat, lng = :lng, phone = :phone
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, library)
	if err != nil {
		return nil, fmt.Errorf("libraryRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("libraryRepo.Update rows error: %w", err)
		}
		return nil, fmt.Errorf("libraryRepo.Update: no rows returned")
	}
	if err := rows.StructScan(library); err != nil {
		return nil, fmt.Errorf("libraryRepo.Update scan: %w", err)
	}
	return library, nil
}

func (r *libraryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM libraries WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("libraryRepo.Delete: %w", err)
	}
	return nil
}
