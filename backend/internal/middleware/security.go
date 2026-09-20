package middleware

import (
	"net/http"
	"strings"

	"dcisp/backend/internal/config"
	"github.com/gin-gonic/gin"
)

// Menerapkan kebijakan CORS berbasis allowlist origin tepercaya dari konfigurasi
// APP_ALLOWED_ORIGINS (SECURITY.md Section 4.3). Kredensial hanya diizinkan untuk
// origin yang cocok persis; origin tak dikenal tidak menerima header ACAO (deny).
// Request tanpa header Origin (non-browser, curl, health check) diteruskan tanpa header CORS.
func CORS(cfg *config.Config) gin.HandlerFunc {
	allowed := make(map[string]bool, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowed[strings.TrimSpace(origin)] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Vary", "Origin")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Device-ID, X-DCISP-Signature, X-Timestamp")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == http.MethodOptions {
			if origin != "" && !allowed[origin] {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Menyematkan security headers HTTP pada setiap respons (SECURITY.md Section 4.4).
// Strict-Transport-Security hanya diaktifkan pada mode release agar tidak mengganggu
// pengembangan lokal yang berjalan di atas HTTP murni.
func SecurityHeaders(cfg *config.Config) gin.HandlerFunc {
	hstsEnabled := strings.EqualFold(cfg.GinMode, "release")

	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https://*.r2.cloudflarestorage.com;")
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if hstsEnabled {
			header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
