package middleware

import (
	"log"
	"runtime/debug"

	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v\nStack: %s", err, string(debug.Stack()))
				response.InternalError(c, "An unexpected internal server error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}
