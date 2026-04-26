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

type libraryRepo struct {
	db *sqlx.DB
}

func NewLibraryRepo(db *sqlx.DB) *libraryRepo {
	return &libraryRepo{db: db}
}

func (r *libraryRepo) Create(ctx context.Context, library *entity.Library) (*entity.Library, error) {
	query := `
        INSERT INTO libraries (id, name, address, lat, lng, phone, photo_url)
        VALUES (:id, :name, :address, :lat, :lng, :phone, :photo_url)
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
	err := r.db.GetContext(ctx, &library,
		`SELECT * FROM libraries WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("libraryRepo.GetByID: %w", err)
	}
	return &library, nil
}

func (r *libraryRepo) List(ctx context.Context, limit, offset int) ([]*entity.Library, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM libraries WHERE deleted_at IS NULL`,
	); err != nil {
		return nil, 0, fmt.Errorf("libraryRepo.List count: %w", err)
	}

	var libraries []*entity.Library
	if err := r.db.SelectContext(ctx, &libraries, `
        SELECT * FROM libraries
        WHERE deleted_at IS NULL
        ORDER BY name
        LIMIT $1 OFFSET $2`,
		limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("libraryRepo.List: %w", err)
	}

	return libraries, total, nil
}

func (r *libraryRepo) ListNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Library, error) {
	var libraries []*entity.Library
	if err := r.db.SelectContext(ctx, &libraries, `
        SELECT * FROM libraries
        WHERE deleted_at IS NULL
          AND (6371 * acos(LEAST(1.0,
                cos(radians($1)) * cos(radians(lat)) *
                cos(radians(lng) - radians($2)) +
                sin(radians($1)) * sin(radians(lat))
              ))) < $3
        ORDER BY name`,
		lat, lng, radiusKm,
	); err != nil {
		return nil, fmt.Errorf("libraryRepo.ListNearby: %w", err)
	}
	return libraries, nil
}

func (r *libraryRepo) Update(ctx context.Context, library *entity.Library) (*entity.Library, error) {
	query := `
        UPDATE libraries
        SET name = :name, address = :address, lat = :lat, lng = :lng, phone = :phone, photo_url = :photo_url
        WHERE id = :id AND deleted_at IS NULL
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
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(library); err != nil {
		return nil, fmt.Errorf("libraryRepo.Update scan: %w", err)
	}
	return library, nil
}

// Delete — soft delete.
// После удаления библиотеки staff теряет доступ автоматически —
// middleware проверяет deleted_at через JOIN при каждом запросе.
func (r *libraryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE libraries SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("libraryRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("libraryRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}
