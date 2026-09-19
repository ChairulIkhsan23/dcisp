package middleware

import (
	"context"
	"strings"

	"dcisp/backend/internal/database"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequirePermission(db *database.PostgresDB, resource, action string, requiredScopeTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get(ContextUserIDKey)
		if !exists {
			response.Unauthorized(c, "Unauthenticated user context")
			c.Abort()
			return
		}

		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			response.Unauthorized(c, "Invalid user identifier")
			c.Abort()
			return
		}

		roleVal, _ := c.Get(ContextRoleKey)
		userRole, _ := roleVal.(string)

		// Enforcement BR-002: Scanner Operator Isolation
		if strings.EqualFold(userRole, "SCANNER_OPERATOR") {
			if !strings.HasPrefix(resource, "attendance.scanner") {
				response.Forbidden(c, "Access Denied: Scanner Operator role is strictly isolated to attendance scanner operations (BR-002)")
				c.Abort()
				return
			}
		}

		// Check permissions dynamically in database (BR-001: Zero Hardcoding)
		ctx := context.Background()
		query := `
			SELECT p.resource, p.action, p.scope_type
			FROM permissions p
			JOIN user_roles ur ON ur.role_id = p.role_id
			WHERE ur.user_id = $1
		`
		rows, err := db.Pool.Query(ctx, query, userID)
		if err != nil {
			response.InternalError(c, "Failed to evaluate authorization permissions", err.Error())
			c.Abort()
			return
		}
		defer rows.Close()

		hasAccess := false
		for rows.Next() {
			var permResource, permAction, permScope string
			if err := rows.Scan(&permResource, &permAction, &permScope); err != nil {
				continue
			}

			// Wildcard check (Super Admin: *.* on SYSTEM)
			if permResource == "*" && permAction == "*" {
				hasAccess = true
				break
			}

			// Resource & Action Match
			if (permResource == resource || permResource == "*") && (permAction == action || permAction == "*") {
				// Scope check
				if len(requiredScopeTypes) == 0 {
					hasAccess = true
					break
				}
				for _, reqScope := range requiredScopeTypes {
					if strings.EqualFold(permScope, reqScope) || strings.EqualFold(permScope, "SYSTEM") {
						hasAccess = true
						break
					}
				}
			}
			if hasAccess {
				break
			}
		}

		if !hasAccess {
			response.Forbidden(c, "Access Denied: You do not have the required permission for this resource/action")
			c.Abort()
			return
		}

		c.Next()
	}
}
