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

type reservationRepo struct {
	db *sqlx.DB
}

func NewReservationRepo(db *sqlx.DB) *reservationRepo {
	return &reservationRepo{db: db}
}

func (r *reservationRepo) CreateWithDecrement(ctx context.Context, res *entity.Reservation) (*entity.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement begin tx: %w", err)
	}
	defer tx.Rollback()

	// ДОБАВИТЬ: проверка дубля через library_book
	if res.LibraryBookID != nil {
		var count int
		if err := tx.GetContext(ctx, &count, `
			SELECT COUNT(*) FROM reservations r
			JOIN library_books lb ON lb.id = r.library_book_id
			WHERE r.user_id = $1
			  AND lb.book_id = (SELECT book_id FROM library_books WHERE id = $2)
			  AND r.status IN ('pending', 'active')`,
			res.UserID, *res.LibraryBookID,
		); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement check duplicate library: %w", err)
		}
		if count > 0 {
			return nil, entity.ErrDuplicateReservation
		}
	}

	// ДОБАВИТЬ: проверка дубля через machine_book
	if res.MachineBookID != nil {
		var count int
		if err := tx.GetContext(ctx, &count, `
			SELECT COUNT(*) FROM reservations r
			JOIN book_machine_books mb ON mb.id = r.machine_book_id
			WHERE r.user_id = $1
			  AND mb.book_id = (SELECT book_id FROM book_machine_books WHERE id = $2)
			  AND r.status IN ('pending', 'active')`,
			res.UserID, *res.MachineBookID,
		); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement check duplicate machine: %w", err)
		}
		if count > 0 {
			return nil, entity.ErrDuplicateReservation
		}
	}

	query := `
		INSERT INTO reservations
			(id, user_id, library_book_id, machine_book_id, source_type, status, reserved_at, due_date, returned_at)
		VALUES
			(:id, :user_id, :library_book_id, :machine_book_id, :source_type, :status, :reserved_at, :due_date, :returned_at)
		RETURNING *`

	rows, err := sqlx.NamedQueryContext(ctx, tx, query, res)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement insert: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement rows error: %w", err)
		}
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement: no rows returned after insert")
	}
	if err := rows.StructScan(res); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement scan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement commit: %w", err)
	}

	return res, nil
}

func (r *reservationRepo) closeAndIncrement(
	ctx context.Context,
	id uuid.UUID,
	callerID *uuid.UUID, // nil = не проверять (admin Return)
	targetStatus entity.ReservationStatus,
	allowedStatuses []entity.ReservationStatus,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement begin tx: %w", err)
	}
	defer tx.Rollback()

	var res entity.Reservation
	if err := tx.GetContext(ctx, &res,
		`SELECT * FROM reservations WHERE id = $1 FOR UPDATE`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrReservationNotFound
		}
		return fmt.Errorf("reservationRepo.closeAndIncrement get reservation: %w", err)
	}

	// ДОБАВИТЬ: проверка владельца внутри транзакции
	if callerID != nil && res.UserID != *callerID {
		return fmt.Errorf("%w: reservation belongs to another user", entity.ErrForbidden)
	}

	if !containsStatus(allowedStatuses, res.Status) {
		return fmt.Errorf("%w: from '%s' to '%s'",
			entity.ErrInvalidStatusTransition, res.Status, targetStatus)
	}

	if targetStatus == entity.ReservationCompleted {
		if _, err := tx.ExecContext(ctx,
			`UPDATE reservations SET status = $1, returned_at = now() WHERE id = $2`,
			targetStatus, id,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement update status: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx,
			`UPDATE reservations SET status = $1 WHERE id = $2`,
			targetStatus, id,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement update status: %w", err)
		}
	}

	if res.LibraryBookID != nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE library_books SET available_copies = available_copies + 1 WHERE id = $1`,
			*res.LibraryBookID,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement increment library: %w", err)
		}
	} else if res.MachineBookID != nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE book_machine_books SET available_copies = available_copies + 1 WHERE id = $1`,
			*res.MachineBookID,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement increment machine: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement commit: %w", err)
	}

	return nil
}

func (r *reservationRepo) ReturnWithIncrement(ctx context.Context, id uuid.UUID) error {
	return r.closeAndIncrement(ctx, id,
		nil, // admin — владельца не проверяем
		entity.ReservationCompleted,
		[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
	)
}

func (r *reservationRepo) CancelWithIncrement(ctx context.Context, id uuid.UUID, callerID *uuid.UUID) error {
	return r.closeAndIncrement(ctx, id,
		callerID,
		entity.ReservationCancelled,
		[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
	)
}
func containsStatus(allowed []entity.ReservationStatus, current entity.ReservationStatus) bool {
	for _, s := range allowed {
		if s == current {
			return true
		}
	}
	return false
}

func (r *reservationRepo) CancelOverdue(ctx context.Context) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("reservationRepo.CancelOverdue begin tx: %w", err)
	}
	defer tx.Rollback()

	var overdue []entity.Reservation
	if err := tx.SelectContext(ctx, &overdue,
		`SELECT * FROM reservations
		 WHERE due_date < now() AND status IN ('active', 'pending')
		 FOR UPDATE`,
	); err != nil {
		return 0, fmt.Errorf("reservationRepo.CancelOverdue select: %w", err)
	}

	if len(overdue) == 0 {
		return 0, nil
	}

	for _, res := range overdue {
		if _, err := tx.ExecContext(ctx,
			`UPDATE reservations SET status = 'cancelled' WHERE id = $1`, res.ID,
		); err != nil {
			return 0, fmt.Errorf("reservationRepo.CancelOverdue update status id=%s: %w", res.ID, err)
		}

		if res.LibraryBookID != nil {
			if _, err := tx.ExecContext(ctx,
				`UPDATE library_books SET available_copies = available_copies + 1 WHERE id = $1`,
				*res.LibraryBookID,
			); err != nil {
				return 0, fmt.Errorf("reservationRepo.CancelOverdue increment library: %w", err)
			}
		} else if res.MachineBookID != nil {
			if _, err := tx.ExecContext(ctx,
				`UPDATE book_machine_books SET available_copies = available_copies + 1 WHERE id = $1`,
				*res.MachineBookID,
			); err != nil {
				return 0, fmt.Errorf("reservationRepo.CancelOverdue increment machine: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("reservationRepo.CancelOverdue commit: %w", err)
	}

	return len(overdue), nil
}

func (r *reservationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	var res entity.Reservation
	if err := r.db.GetContext(ctx, &res, `SELECT * FROM reservations WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, fmt.Errorf("reservationRepo.GetByID: %w", err)
	}
	return &res, nil
}

func (r *reservationRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM reservations WHERE user_id = $1`, userID,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByUser count: %w", err)
	}

	var items []*entity.Reservation
	if err := r.db.SelectContext(ctx, &items,
		`SELECT * FROM reservations WHERE user_id = $1
		 ORDER BY reserved_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByUser: %w", err)
	}

	return items, total, nil
}

func (r *reservationRepo) ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM reservations WHERE ($1::text IS NULL OR status = $1)`, status,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll count: %w", err)
	}

	var items []*entity.Reservation
	if err := r.db.SelectContext(ctx, &items,
		`SELECT * FROM reservations
		 WHERE ($1::text IS NULL OR status = $1)
		 ORDER BY reserved_at DESC LIMIT $2 OFFSET $3`,
		status, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll: %w", err)
	}

	return items, total, nil
}
func (r *reservationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	var (
		result sql.Result
		err    error
	)
	if status == entity.ReservationCompleted {
		result, err = r.db.ExecContext(ctx,
			`UPDATE reservations SET status = $1, returned_at = now() WHERE id = $2`, status, id)
	} else {
		result, err = r.db.ExecContext(ctx,
			`UPDATE reservations SET status = $1 WHERE id = $2`, status, id)
	}
	if err != nil {
		return fmt.Errorf("reservationRepo.UpdateStatus: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reservationRepo.UpdateStatus rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrReservationNotFound
	}
	return nil
}

func (r *reservationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM reservations WHERE id = $1`, id); err != nil {
		return fmt.Errorf("reservationRepo.Delete: %w", err)
	}
	return nil
}
