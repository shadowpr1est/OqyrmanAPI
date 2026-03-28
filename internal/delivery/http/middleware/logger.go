package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs each HTTP request with method, path, status, latency and client IP.
// 5xx → Error, 4xx → Warn, rest → Info.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.String("latency", latency.String()),
			slog.String("ip", c.ClientIP()),
		}

		if uid, exists := c.Get("userID"); exists {
			attrs = append(attrs, slog.Any("user_id", uid))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.LogAttrs(c.Request.Context(), level, "request", attrs...)
	}
}
