package middleware

import (
	"strings"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/shared/response"
	"dcisp/backend/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey = "user_id"
	ContextEmailKey  = "email"
	ContextRoleKey   = "role"
	ContextScopesKey = "scopes"
)

// Memvalidasi token JWT Bearer pada header request dan menyematkan data identitas pengguna ke dalam request context.
func AuthJWT(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header missing")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Invalid authorization header format. Expected 'Bearer <token>'")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(cfg.JWTSecret, parts[1])
		if err != nil {
			response.Unauthorized(c, "Invalid or expired access token")
			c.Abort()
			return
		}

		// Inject user context
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextEmailKey, claims.Email)
		c.Set(ContextRoleKey, claims.Role)
		c.Set(ContextScopesKey, claims.Scopes)

		c.Next()
	}
}
