package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	secretKey      string
	accessTokenTTL time.Duration
}

type Claims struct {
	UserID    string     `json:"user_id"`
	Role      string     `json:"role"`
	LibraryID *uuid.UUID `json:"library_id,omitempty"` // ← НОВОЕ
	jwt.RegisteredClaims
}

func NewManager(secretKey string, accessTTLMin int) (*Manager, error) {
	if len([]byte(secretKey)) < 32 {
		return nil, errors.New("jwt secret key must be at least 32 bytes")
	}
	return &Manager{
		secretKey:      secretKey,
		accessTokenTTL: time.Duration(accessTTLMin) * time.Minute,
	}, nil
}

func (m *Manager) GenerateAccessToken(userID uuid.UUID, role string, libraryID *uuid.UUID) (string, error) {
	claims := &Claims{
		UserID:    userID.String(),
		Role:      role,
		LibraryID: libraryID, // ← НОВОЕ
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
