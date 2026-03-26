package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
)

const (
	UserIDKey    = "userID"
	RoleKey      = "role"
	LibraryIDKey = "libraryID" // ← НОВОЕ
)

func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := jwtManager.ParseAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		// ❗ Валидация staff: library_id обязателен
		if claims.Role == "Staff" && claims.LibraryID == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "staff must have library_id"})
			return
		}

		c.Set(UserIDKey, userID)
		c.Set(RoleKey, claims.Role)
		c.Set(LibraryIDKey, claims.LibraryID) // *uuid.UUID, может быть nil
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(RoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if role != "Admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func LibraryStaffOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(RoleKey)
		if role != "Admin" && role != "Staff" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "staff access required"})
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) uuid.UUID {
	return c.MustGet(UserIDKey).(uuid.UUID)
}

func GetLibraryID(c *gin.Context) *uuid.UUID {
	val, exists := c.Get(LibraryIDKey)
	if !exists {
		return nil
	}
	id, _ := val.(*uuid.UUID)
	return id
}
