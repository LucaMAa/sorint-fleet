package middleware

import (
	"strings"

	"sorint-fleet/internal/config"
	"sorint-fleet/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "user_id"
	ContextRole   = "user_role"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "missing token")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Authorization not valid (expected: Bearer <token>)")
			c.Abort()
			return
		}

		claims, err := config.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "token not valid or expired")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextRole)
		if !exists {
			response.Unauthorized(c, "role not found in token")
			c.Abort()
			return
		}
		if _, ok := allowed[role.(string)]; !ok {
			response.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
