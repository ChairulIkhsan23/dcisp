package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"dcisp/backend/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuditInterceptor(db *database.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only audit state-changing operations or sensitive endpoints
		method := c.Request.Method
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
			path := c.Request.URL.Path
			// Skip health endpoints
			if strings.HasPrefix(path, "/health") {
				return
			}

			var actorID *uuid.UUID
			if val, exists := c.Get(ContextUserIDKey); exists {
				if uid, ok := val.(uuid.UUID); ok {
					actorID = &uid
				}
			}

			ip := c.ClientIP()
			ua := c.Request.UserAgent()
			eventName := method + " " + path

			// Insert asynchronously to prevent blocking response thread
			go func(uid *uuid.UUID, event, ipAddr, userAgent string) {
				query := `
					INSERT INTO audit_logs (id, user_id, event_name, resource_type, resource_id, ip_address, user_agent, timestamp)
					VALUES (gen_random_uuid(), $1, $2, 'API_ENDPOINT', $3, $4, $5, CURRENT_TIMESTAMP)
				`
				_, err := db.Pool.Exec(context.Background(), query, uid, event, path, ipAddr, userAgent)
				if err != nil {
					log.Printf("Failed to record audit log: %v", err)
				}
			}(actorID, eventName, ip, ua)
		}
	}
}
