package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

	// Проверка дубля: нельзя бронировать одну книгу дважды
	var count int
	if err := tx.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM reservations r
		JOIN library_books lb ON lb.id = r.library_book_id
		WHERE r.user_id = $1
		  AND lb.book_id = (SELECT book_id FROM library_books WHERE id = $2)
		  AND r.status IN ('pending', 'active')`,
		res.UserID, res.LibraryBookID,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement check duplicate: %w", err)
	}
	if count > 0 {
		return nil, entity.ErrDuplicateReservation
	}

	// Декремент available_copies с проверкой наличия
	var available int
	if err := tx.GetContext(ctx, &available,
		`SELECT available_copies FROM library_books WHERE id = $1 FOR UPDATE`,
		res.LibraryBookID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement get copies: %w", err)
	}
	if available <= 0 {
		return nil, entity.ErrNoAvailableCopies
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE library_books SET available_copies = available_copies - 1 WHERE id = $1`,
		res.LibraryBookID,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement decrement: %w", err)
	}

	query := `
		INSERT INTO reservations (id, user_id, library_book_id, status, reserved_at, due_date, returned_at)
		VALUES (:id, :user_id, :library_book_id, :status, :reserved_at, :due_date, :returned_at)
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
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement: no rows returned")
	}
	if err := rows.StructScan(res); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement scan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement commit: %w", err)
	}

	return res, nil
}

// closeAndIncrement — общая логика для cancel и return.
// callerID != nil → проверяет что бронь принадлежит пользователю
// libraryID != nil → проверяет что бронь принадлежит библиотеке staff
// оба nil → admin, без проверок
func (r *reservationRepo) closeAndIncrement(
	ctx context.Context,
	id uuid.UUID,
	callerID *uuid.UUID,
	libraryID *uuid.UUID,
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
		return fmt.Errorf("reservationRepo.closeAndIncrement get: %w", err)
	}

	// Проверка владельца — для отмены пользователем
	if callerID != nil && res.UserID != *callerID {
		return fmt.Errorf("%w: reservation belongs to another user", entity.ErrForbidden)
	}

	// Проверка принадлежности библиотеке — для staff операций
	if libraryID != nil {
		var belongs bool
		if err := tx.GetContext(ctx, &belongs, `
			SELECT EXISTS(
				SELECT 1 FROM library_books lb
				WHERE lb.id = $1 AND lb.library_id = $2
			)`, res.LibraryBookID, *libraryID,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement check library: %w", err)
		}
		if !belongs {
			return fmt.Errorf("%w: reservation belongs to another library", entity.ErrForbidden)
		}
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

	if _, err := tx.ExecContext(ctx,
		`UPDATE library_books SET available_copies = available_copies + 1 WHERE id = $1`,
		res.LibraryBookID,
	); err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement increment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement commit: %w", err)
	}

	return nil
}

func (r *reservationRepo) CancelWithIncrement(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	return r.closeAndIncrement(ctx, id, &callerID, nil,
		entity.ReservationCancelled,
		[]entity.ReservationStatus{entity.ReservationPending},
	)
}

func (r *reservationRepo) ListByLibrary(
	ctx context.Context,
	libraryID uuid.UUID,
	limit, offset int,
	status *string,
) ([]*entity.Reservation, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(res.id)
		FROM reservations res
		JOIN library_books lb ON lb.id = res.library_book_id
		WHERE lb.library_id = $1
		  AND ($2::text IS NULL OR res.status = $2)`,
		libraryID, status,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByLibrary count: %w", err)
	}

	var items []*entity.Reservation
	if err := r.db.SelectContext(ctx, &items, `
		SELECT res.*
		FROM reservations res
		JOIN library_books lb ON lb.id = res.library_book_id
		WHERE lb.library_id = $1
		  AND ($2::text IS NULL OR res.status = $2)
		ORDER BY res.reserved_at DESC
		LIMIT $3 OFFSET $4`,
		libraryID, status, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByLibrary: %w", err)
	}

	return items, total, nil
}

func (r *reservationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	var res entity.Reservation
	if err := r.db.GetContext(ctx, &res,
		`SELECT * FROM reservations WHERE id = $1`, id,
	); err != nil {
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
	if err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM reservations
		WHERE user_id = $1
		ORDER BY reserved_at DESC
		LIMIT $2 OFFSET $3`,
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
	if err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM reservations
		WHERE ($1::text IS NULL OR status = $1)
		ORDER BY reserved_at DESC
		LIMIT $2 OFFSET $3`,
		status, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll: %w", err)
	}

	return items, total, nil
}

func (r *reservationRepo) Extend(ctx context.Context, id, userID uuid.UUID, newDueDate time.Time) (*entity.Reservation, error) {
	var res entity.Reservation
	err := r.db.GetContext(ctx, &res, `
		UPDATE reservations
		SET due_date = $3
		WHERE id = $1 AND user_id = $2 AND status = 'active' AND deleted_at IS NULL
		RETURNING *`,
		id, userID, newDueDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, fmt.Errorf("reservationRepo.Extend: %w", err)
	}
	return &res, nil
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

func (r *reservationRepo) CancelOverdue(ctx context.Context) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("reservationRepo.CancelOverdue begin tx: %w", err)
	}
	defer tx.Rollback()

	var overdue []entity.Reservation
	if err := tx.SelectContext(ctx, &overdue, `
		SELECT * FROM reservations
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
			return 0, fmt.Errorf("reservationRepo.CancelOverdue update id=%s: %w", res.ID, err)
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE library_books SET available_copies = available_copies + 1 WHERE id = $1`,
			res.LibraryBookID,
		); err != nil {
			return 0, fmt.Errorf("reservationRepo.CancelOverdue increment id=%s: %w", res.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("reservationRepo.CancelOverdue commit: %w", err)
	}

	return len(overdue), nil
}

func (r *reservationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM reservations WHERE id = $1`, id,
	); err != nil {
		return fmt.Errorf("reservationRepo.Delete: %w", err)
	}
	return nil
}

func (r *reservationRepo) StaffCancel(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	return r.closeAndIncrement(ctx, id, nil, &libraryID,
		entity.ReservationCancelled,
		[]entity.ReservationStatus{entity.ReservationPending},
	)
}

func (r *reservationRepo) StaffReturn(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	return r.closeAndIncrement(ctx, id, nil, &libraryID,
		entity.ReservationCompleted,
		[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
	)
}

func (r *reservationRepo) AdminReturn(ctx context.Context, id uuid.UUID) error {
	return r.closeAndIncrement(ctx, id, nil, nil,
		entity.ReservationCompleted,
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
