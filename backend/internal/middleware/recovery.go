package middleware

import (
	"log"
	"runtime/debug"

	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Menangkap panic pada aplikasi dan mengembalikan respons error 500 tanpa menghentikan server.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v\nStack: %s", err, string(debug.Stack()))
				response.InternalError(c, "Terjadi kesalahan internal server yang tidak terduga")
				c.Abort()
			}
		}()
		c.Next()
	}
}
