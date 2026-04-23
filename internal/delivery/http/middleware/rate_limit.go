package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	window   time.Duration
}

func NewRateLimiter(ctx context.Context, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
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

func (rl *RateLimiter) allow(key string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]

	if !exists || now.Sub(v.windowStart) >= rl.window {
		rl.visitors[key] = &visitor{count: 1, windowStart: now}
		return true
	}

	if v.count >= limit {
		return false
	}

	v.count++
	return true
}

func RateLimitWithGroup(rl *RateLimiter, group string, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() + ":" + group
		if !rl.allow(key, limit) {
			retryAfter := int(rl.window.Seconds())
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    "too_many_requests",
				"message": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}

// RateLimitPerUser ограничивает по userID из JWT (а не по IP). Используется
// для дорогих эндпойнтов вроде LLM-стриминга, где один пользователь по NAT'у
// не должен расходовать общий бюджет с соседями, и наоборот — один юзер с
// нескольких IP должен укладываться в свою квоту.
func RateLimitPerUser(rl *RateLimiter, group string, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, exists := c.Get(UserIDKey)
		if !exists {
			c.Next()
			return
		}
		key := fmt.Sprintf("u:%v:%s", uid, group)
		if !rl.allow(key, limit) {
			retryAfter := int(rl.window.Seconds())
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    "too_many_requests",
				"message": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
