package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"dcisp/backend/internal/database"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CachedPermission struct {
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	ScopeType string `json:"scope_type"`
}

// Mengevaluasi hak akses dan batasan scope pengguna secara dinamis dengan caching Redis 5 menit.
func RequirePermission(db *database.PostgresDB, rdb *database.RedisClient, resource, action string, requiredScopeTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get(ContextUserIDKey)
		if !exists {
			response.Unauthorized(c, "Konteks pengguna tidak terautentikasi")
			c.Abort()
			return
		}

		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			response.Unauthorized(c, "Pengenal pengguna tidak valid")
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		cacheKey := fmt.Sprintf("rbac:permissions:%s", userID.String())

		var permissions []CachedPermission
		cacheHit := false

		// 1. Cek Redis Cache jika client Redis tersedia
		if rdb != nil && rdb.Client != nil {
			cachedData, err := rdb.Client.Get(ctx, cacheKey).Result()
			if err == nil {
				if err := json.Unmarshal([]byte(cachedData), &permissions); err == nil {
					cacheHit = true
				}
			} else if err != redis.Nil {
				log.Printf("Peringatan: Gagal membaca cache RBAC dari Redis: %v", err)
				response.InternalError(c, "Terjadi kesalahan pada sistem otorisasi")
				c.Abort()
				return
			}
		}

		// 2. Cache Miss: Ambil dari database PostgreSQL (BR-001: Zero Hardcoding)
		if !cacheHit {
			query := `
				SELECT p.resource, p.action, p.scope_type
				FROM permissions p
				JOIN user_roles ur ON ur.role_id = p.role_id
				WHERE ur.user_id = $1
			`
			rows, err := db.Pool.Query(ctx, query, userID)
			if err != nil {
				log.Printf("Kesalahan kueri izin RBAC dari database: %v", err)
				response.InternalError(c, "Terjadi kesalahan pada sistem otorisasi")
				c.Abort()
				return
			}
			defer rows.Close()

			for rows.Next() {
				var perm CachedPermission
				if err := rows.Scan(&perm.Resource, &perm.Action, &perm.ScopeType); err == nil {
					permissions = append(permissions, perm)
				}
			}

			// Simpan ke Redis dengan TTL 5 menit
			if rdb != nil && rdb.Client != nil {
				if permBytes, err := json.Marshal(permissions); err == nil {
					if setErr := rdb.Client.Set(ctx, cacheKey, permBytes, 5*time.Minute).Err(); setErr != nil {
						log.Printf("Peringatan: Gagal menyimpan cache RBAC ke Redis: %v", setErr)
					}
				}
			}
		}

		// 3. Evaluasi Izin Pengguna
		hasAccess := false
		for _, perm := range permissions {
			// Wildcard check (Super Admin: *.* on SYSTEM)
			if perm.Resource == "*" && perm.Action == "*" {
				hasAccess = true
				break
			}

			// Resource & Action Match
			if (perm.Resource == resource || perm.Resource == "*") && (perm.Action == action || perm.Action == "*") {
				// Scope check
				if len(requiredScopeTypes) == 0 {
					hasAccess = true
					break
				}
				for _, reqScope := range requiredScopeTypes {
					if strings.EqualFold(perm.ScopeType, reqScope) || strings.EqualFold(perm.ScopeType, "SYSTEM") {
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
			response.Forbidden(c, "Akses Ditolak: Anda tidak memiliki izin yang diperlukan untuk sumber daya/tindakan ini")
			c.Abort()
			return
		}

		c.Next()
	}
}
