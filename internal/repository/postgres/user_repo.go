package postgres

import (
	"context"
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
		INSERT INTO users (id, email, phone, password_hash, full_name, avatar_url, role, qr_code, created_at)
		VALUES (:id, :email, :phone, :password_hash, :full_name, :avatar_url, :role, :qr_code, :created_at)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("userRepo.Create: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(user); err != nil {
		return nil, fmt.Errorf("userRepo.Create scan: %w", err)
	}
	return user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	query := `SELECT * FROM users WHERE id = $1`
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		return nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	query := `SELECT * FROM users WHERE email = $1`
	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		return nil, fmt.Errorf("userRepo.GetByEmail: %w", err)
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	// role and qr_code are intentionally excluded — they must not be changed via user profile update
	query := `
		UPDATE users
		SET email = :email, phone = :phone, full_name = :full_name, avatar_url = :avatar_url
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("userRepo.Update: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(user); err != nil {
		return nil, fmt.Errorf("userRepo.Update scan: %w", err)
	}
	return user, nil
}

func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("userRepo.Delete: %w", err)
	}
	return nil
}

func (r *userRepo) ListAll(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	var users []*entity.User
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users`); err != nil {
		return nil, 0, fmt.Errorf("userRepo.ListAll count: %w", err)
	}
	if err := r.db.SelectContext(ctx, &users, `SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("userRepo.ListAll: %w", err)
	}
	return users, total, nil
}

func (r *userRepo) UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET avatar_url = $1 WHERE id = $2`, avatarURL, id)
	return err
}

func (r *userRepo) UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, id)
	return err
}
