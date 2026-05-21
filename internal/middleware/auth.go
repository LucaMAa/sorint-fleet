package middleware

import (
	"sorint-fleet/internal/session"
	"sorint-fleet/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ContextUserID    = "user_id"
	ContextRole      = "user_role"
	ContextSessionID = "session_id"
)

func Auth(store *session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(session.CookieName)
		if err != nil || sessionID == "" {
			response.Unauthorized(c, "missing session")
			c.Abort()
			return
		}

		data, err := store.Get(c.Request.Context(), sessionID)
		if err != nil {
			response.Unauthorized(c, "session invalid or expired")
			c.Abort()
			return
		}

		c.Set(ContextUserID, data.UserID)
		c.Set(ContextRole, data.Role)
		c.Set(ContextSessionID, sessionID)
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
			response.Unauthorized(c, "role not found in session")
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

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, _ := c.Get(ContextUserID)
	uid, ok := raw.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "not authenticated")
	}
	return uid, ok
}

func GetSessionID(c *gin.Context) string {
	raw, _ := c.Get(ContextSessionID)
	sid, _ := raw.(string)
	return sid
}
