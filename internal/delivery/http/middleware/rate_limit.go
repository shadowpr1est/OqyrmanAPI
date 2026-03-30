package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count       int
	windowStart time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func newRateLimiter(ctx context.Context, limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	// чистим старые записи каждые 5 минут; горутина живёт ровно столько, сколько ctx
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rl.mu.Lock()
				for ip, v := range rl.visitors {
					if time.Since(v.windowStart) > rl.window {
						delete(rl.visitors, ip)
					}
				}
				rl.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok || time.Since(v.windowStart) > rl.window {
		rl.visitors[ip] = &visitor{count: 1, windowStart: time.Now()}
		return true
	}

	if v.count >= rl.limit {
		return false
	}

	v.count++
	return true
}

// RateLimit — универсальный middleware: limit запросов за window.
// ctx должен быть контекстом приложения — при его отмене cleanup-горутина завершается.
func RateLimit(ctx context.Context, limit int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(ctx, limit, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
