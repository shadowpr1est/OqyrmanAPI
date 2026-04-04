package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/ctxkeys"
)

// InjectRequestMeta injects User-Agent and client IP into the request context
// so that usecases can access them without depending on gin.
func InjectRequestMeta() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ctxkeys.UserAgentKey, c.Request.UserAgent())
		ctx = context.WithValue(ctx, ctxkeys.ClientIPKey, c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
