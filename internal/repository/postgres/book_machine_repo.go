package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type bookMachineRepo struct {
	db *sqlx.DB
}

func NewBookMachineRepo(db *sqlx.DB) *bookMachineRepo {
	return &bookMachineRepo{db: db}
}

func (r *bookMachineRepo) Create(ctx context.Context, machine *entity.BookMachine) (*entity.BookMachine, error) {
	query := `
		INSERT INTO book_machines (id, name, address, lat, lng, status)
		VALUES (:id, :name, :address, :lat, :lng, :status)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, machine)
	if err != nil {
		return nil, fmt.Errorf("bookMachineRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("bookMachineRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("bookMachineRepo.Create: no rows returned")
	}
	if err := rows.StructScan(machine); err != nil {
		return nil, fmt.Errorf("bookMachineRepo.Create scan: %w", err)
	}
	return machine, nil
}

func (r *bookMachineRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.BookMachine, error) {
	var machine entity.BookMachine
	query := `SELECT * FROM book_machines WHERE id = $1`
	if err := r.db.GetContext(ctx, &machine, query, id); err != nil {
		return nil, fmt.Errorf("bookMachineRepo.GetByID: %w", err)
	}
	return &machine, nil
}

func (r *bookMachineRepo) List(ctx context.Context, limit, offset int) ([]*entity.BookMachine, int, error) {
	var machines []*entity.BookMachine
	var total int
	query := `SELECT * FROM book_machines ORDER BY name LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &machines, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("bookMachineRepo.List: %w", err)
	}
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM book_machines`); err != nil {
		return nil, 0, fmt.Errorf("bookMachineRepo.List count: %w", err)
	}
	return machines, total, nil
}

func (r *bookMachineRepo) ListNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.BookMachine, error) {
	var machines []*entity.BookMachine
	// LEAST(1.0, ...) защищает acos от NaN при точном совпадении координат
	query := `
		SELECT * FROM book_machines
		WHERE (6371 * acos(LEAST(1.0, cos(radians($1)) * cos(radians(lat)) *
		cos(radians(lng) - radians($2)) + sin(radians($1)) * sin(radians(lat))))) < $3
		ORDER BY name`
	if err := r.db.SelectContext(ctx, &machines, query, lat, lng, radiusKm); err != nil {
		return nil, fmt.Errorf("bookMachineRepo.ListNearby: %w", err)
	}
	return machines, nil
}

func (r *bookMachineRepo) Update(ctx context.Context, machine *entity.BookMachine) (*entity.BookMachine, error) {
	query := `
		UPDATE book_machines
		SET name = :name, address = :address, lat = :lat, lng = :lng, status = :status
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, machine)
	if err != nil {
		return nil, fmt.Errorf("bookMachineRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("bookMachineRepo.Update rows error: %w", err)
		}
		return nil, fmt.Errorf("bookMachineRepo.Update: no rows returned")
	}
	if err := rows.StructScan(machine); err != nil {
		return nil, fmt.Errorf("bookMachineRepo.Update scan: %w", err)
	}
	return machine, nil
}

func (r *bookMachineRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM book_machines WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("bookMachineRepo.Delete: %w", err)
	}
	return nil
}
