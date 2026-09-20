package system

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/shared/eventbus"
	"dcisp/backend/internal/shared/response"
	"dcisp/backend/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type SSEController struct {
	cfg      *config.Config
	eventBus *eventbus.EventBus
}

// Menginisialisasi instance baru controller Server-Sent Events (SSE).
func NewSSEController(cfg *config.Config, bus *eventbus.EventBus) *SSEController {
	return &SSEController{cfg: cfg, eventBus: bus}
}

// Menangani koneksi streaming Server-Sent Events (SSE) untuk siaran peristiwa real-time.
func (ctrl *SSEController) StreamEvents(c *gin.Context) {
	// Autentikasi: Dukung Header Authorization atau Query Param token (karena batasan EventSource browser)
	tokenStr := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			tokenStr = parts[1]
		}
	}
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}

	if tokenStr == "" {
		response.Unauthorized(c, "Token autentikasi SSE tidak ditemukan")
		return
	}

	claims, err := utils.ValidateToken(ctrl.cfg.JWTSecret, tokenStr, utils.TokenTypeAccess)
	if err != nil {
		response.Unauthorized(c, "Token akses SSE tidak valid atau telah kedaluwarsa")
		return
	}

	// Setup SSE Headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Flush()

	clientChan := make(chan eventbus.DomainEvent, 20)

	// Subscribe ke EventBus
	ctrl.eventBus.Subscribe("*", func(ctx context.Context, event eventbus.DomainEvent) error {
		// Multi-tenant & User isolation: kirim jika event publik atau milik user yang sedang terhubung
		if targetUser, ok := event.Payload["user_id"].(string); ok && targetUser != "" {
			if targetUser != claims.UserID.String() {
				return nil
			}
		}
		select {
		case clientChan <- event:
		default:
			// Buffer penuh, abaikan event tanpa memblokir
		}
		return nil
	})

	// Kirim pesan koneksi terhubung awal
	initialEvent, _ := json.Marshal(gin.H{
		"connected": true,
		"user_id":   claims.UserID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	fmt.Fprintf(c.Writer, "event: connected\ndata: %s\n\n", initialEvent)
	c.Writer.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case <-ticker.C:
			// Detak jantung keepalive berkala
			fmt.Fprintf(w, ": keepalive\n\n")
			return true
		case event := <-clientChan:
			eventData, err := json.Marshal(event.Payload)
			if err != nil {
				return true
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, eventData)
			return true
		}
	})
}
