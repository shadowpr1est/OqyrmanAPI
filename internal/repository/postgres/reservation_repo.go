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

// CreateWithDecrement атомарно:
// 1. Блокирует строку копий (FOR UPDATE) — исключает race condition
// 2. Проверяет available_copies > 0
// 3. Уменьшает available_copies на 1
// 4. Создаёт запись бронирования
func (r *reservationRepo) CreateWithDecrement(ctx context.Context, res *entity.Reservation) (*entity.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement begin tx: %w", err)
	}
	defer tx.Rollback()

	if res.LibraryBookID != nil {
		var available int
		err := tx.GetContext(ctx, &available,
			`SELECT available_copies FROM library_books WHERE id = $1 FOR UPDATE`,
			*res.LibraryBookID,
		)
		if err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement get library_book: %w", err)
		}
		if available <= 0 {
			return nil, fmt.Errorf("no available copies in library")
		}
		if _, err = tx.ExecContext(ctx,
			`UPDATE library_books SET available_copies = available_copies - 1 WHERE id = $1`,
			*res.LibraryBookID,
		); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement decrement library: %w", err)
		}
	}

	if res.MachineBookID != nil {
		var available int
		err := tx.GetContext(ctx, &available,
			`SELECT available_copies FROM book_machine_books WHERE id = $1 FOR UPDATE`,
			*res.MachineBookID,
		)
		if err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement get machine_book: %w", err)
		}
		if available <= 0 {
			return nil, fmt.Errorf("no available copies in book machine")
		}
		if _, err = tx.ExecContext(ctx,
			`UPDATE book_machine_books SET available_copies = available_copies - 1 WHERE id = $1`,
			*res.MachineBookID,
		); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement decrement machine: %w", err)
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

// closeAndIncrement — приватный хелпер для атомарного закрытия брони.
func (r *reservationRepo) closeAndIncrement(
	ctx context.Context,
	id uuid.UUID,
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
		return fmt.Errorf("reservationRepo.closeAndIncrement get reservation: %w", err)
	}

	if !containsStatus(allowedStatuses, res.Status) {
		return fmt.Errorf("cannot transition reservation from '%s' to '%s'",
			res.Status, targetStatus)
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

func containsStatus(allowed []entity.ReservationStatus, current entity.ReservationStatus) bool {
	for _, s := range allowed {
		if s == current {
			return true
		}
	}
	return false
}

func (r *reservationRepo) ReturnWithIncrement(ctx context.Context, id uuid.UUID) error {
	return r.closeAndIncrement(ctx, id,
		entity.ReservationCompleted,
		[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
	)
}

func (r *reservationRepo) CancelWithIncrement(ctx context.Context, id uuid.UUID) error {
	return r.closeAndIncrement(ctx, id,
		entity.ReservationCancelled,
		[]entity.ReservationStatus{entity.ReservationPending},
	)
}

// CancelOverdue в одной транзакции:
// 1. Блокирует все активные просроченные брони (FOR UPDATE)
// 2. Меняет статус на cancelled
// 3. Восстанавливает available_copies для каждой брони
func (r *reservationRepo) CancelOverdue(ctx context.Context) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("reservationRepo.CancelOverdue begin tx: %w", err)
	}
	defer tx.Rollback()

	var overdue []entity.Reservation
	if err := tx.SelectContext(ctx, &overdue,
		`SELECT * FROM reservations
		 WHERE due_date < now() AND status = 'active'
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
				return 0, fmt.Errorf("reservationRepo.CancelOverdue increment library id=%s: %w", *res.LibraryBookID, err)
			}
		} else if res.MachineBookID != nil {
			if _, err := tx.ExecContext(ctx,
				`UPDATE book_machine_books SET available_copies = available_copies + 1 WHERE id = $1`,
				*res.MachineBookID,
			); err != nil {
				return 0, fmt.Errorf("reservationRepo.CancelOverdue increment machine id=%s: %w", *res.MachineBookID, err)
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
	query := `SELECT * FROM reservations WHERE id = $1`
	if err := r.db.GetContext(ctx, &res, query, id); err != nil {
		return nil, fmt.Errorf("reservationRepo.GetByID: %w", err)
	}
	return &res, nil
}

// ListByUser возвращает брони пользователя с пагинацией и общим количеством.
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
		 ORDER BY reserved_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByUser: %w", err)
	}

	return items, total, nil
}

func (r *reservationRepo) ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	var items []*entity.Reservation
	var total int

	countQuery := `SELECT COUNT(*) FROM reservations WHERE ($1::text IS NULL OR status = $1)`
	if err := r.db.GetContext(ctx, &total, countQuery, status); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll count: %w", err)
	}

	query := `
		SELECT * FROM reservations
		WHERE ($1::text IS NULL OR status = $1)
		ORDER BY reserved_at DESC
		LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &items, query, status, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll: %w", err)
	}

	return items, total, nil
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
	if err != nil {
		return fmt.Errorf("reservationRepo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *reservationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reservations WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("reservationRepo.Delete: %w", err)
	}
	return nil
}
