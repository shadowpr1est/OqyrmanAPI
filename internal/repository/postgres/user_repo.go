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

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := `
		INSERT INTO users (id, email, phone, password_hash, full_name, avatar_url, role, library_id, qr_code, created_at)
		VALUES (:id, :email, :phone, :password_hash, :full_name, :avatar_url, :role, :library_id, :qr_code, :created_at)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("userRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("userRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("userRepo.Create: no rows returned")
	}
	if err := rows.StructScan(user); err != nil {
		return nil, fmt.Errorf("userRepo.Create scan: %w", err)
	}
	return user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`, email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("userRepo.GetByEmail: %w", err)
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	// role, library_id, qr_code намеренно исключены — не меняются через профиль
	// avatar_url исключён — обновляется только через UpdateAvatarURL
	query := `
		UPDATE users
		SET email = :email, phone = :phone, full_name = :full_name
		WHERE id = :id AND deleted_at IS NULL
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("userRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("userRepo.Update rows error: %w", err)
		}
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(user); err != nil {
		return nil, fmt.Errorf("userRepo.Update scan: %w", err)
	}
	return user, nil
}

// Delete — soft delete.
// Токены не трогаем — они протухнут сами по TTL.
// При следующем запросе GetByEmail/GetByID вернёт ErrNotFound.
func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("userRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("userRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *userRepo) ListAll(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`,
	); err != nil {
		return nil, 0, fmt.Errorf("userRepo.ListAll count: %w", err)
	}

	var users []*entity.User
	if err := r.db.SelectContext(ctx, &users, `
		SELECT * FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("userRepo.ListAll: %w", err)
	}

	return users, total, nil
}

func (r *userRepo) UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET avatar_url = $1 WHERE id = $2 AND deleted_at IS NULL`,
		avatarURL, id,
	)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateAvatarURL: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("userRepo.UpdateAvatarURL rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// UpdateRole — при смене роли на не-Staff, library_id обнуляется.
// Это соответствует constraint chk_staff_library в БД.
func (r *userRepo) UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role, libraryID *uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET role = $1, library_id = $2 WHERE id = $3 AND deleted_at IS NULL`,
		role, libraryID, id,
	)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateRole: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("userRepo.UpdateRole rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}
