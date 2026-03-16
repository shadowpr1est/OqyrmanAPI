package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type reservationRepo struct {
	db *sqlx.DB
}

func NewReservationRepo(db *sqlx.DB) *reservationRepo {
	return &reservationRepo{db: db}
}

func (r *reservationRepo) Create(ctx context.Context, res *entity.Reservation) (*entity.Reservation, error) {
	query := `
		INSERT INTO reservations (id, user_id, library_book_id, machine_book_id, source_type, status, reserved_at, due_date, returned_at)
		VALUES (:id, :user_id, :library_book_id, :machine_book_id, :source_type, :status, :reserved_at, :due_date, :returned_at)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, res)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.Create: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(res); err != nil {
		return nil, fmt.Errorf("reservationRepo.Create scan: %w", err)
	}
	return res, nil
}

func (r *reservationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	var res entity.Reservation
	query := `SELECT * FROM reservations WHERE id = $1`
	if err := r.db.GetContext(ctx, &res, query, id); err != nil {
		return nil, fmt.Errorf("reservationRepo.GetByID: %w", err)
	}
	return &res, nil
}

func (r *reservationRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Reservation, error) {
	var items []*entity.Reservation
	query := `SELECT * FROM reservations WHERE user_id = $1 ORDER BY reserved_at DESC`
	if err := r.db.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, fmt.Errorf("reservationRepo.ListByUser: %w", err)
	}
	return items, nil
}

func (r *reservationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reservations WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("reservationRepo.Delete: %w", err)
	}
	return nil
}

func (r *reservationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	var err error
	if status == entity.ReservationCompleted {
		_, err = r.db.ExecContext(ctx,
			`UPDATE reservations SET status = $1, returned_at = now() WHERE id = $2`, status, id)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE reservations SET status = $1 WHERE id = $2`, status, id)
	}
	return err
}
