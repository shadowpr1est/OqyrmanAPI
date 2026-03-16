package entity

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin Role = "Admin"
	RoleUser  Role = "User"
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	FullName     string    `db:"full_name"`
	Phone        string    `db:"phone"`
	AvatarURL    string    `db:"avatar_url"`
	Role         Role      `db:"role"`
	QRCode       string    `db:"qr_code"`
	CreatedAt    time.Time `db:"created_at"`
}
