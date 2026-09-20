package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"dcisp/backend/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type auditEntry struct {
	ActorID   *uuid.UUID
	EventName string
	Path      string
	IP        string
	UserAgent string
	NewState  []byte
}

var (
	auditQueue  chan auditEntry
	auditWg     sync.WaitGroup
	auditOnce   sync.Once
	auditClosed bool
	auditLock   sync.Mutex
)

// Menginisialisasi worker asinkron untuk mencatat jejak audit ke database secara non-blocking.
func InitAuditWorker(db *database.PostgresDB) {
	auditOnce.Do(func() {
		auditQueue = make(chan auditEntry, 1000)
		auditWg.Add(1)
		go func() {
			defer auditWg.Done()
			for entry := range auditQueue {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				query := `
					INSERT INTO audit_logs (id, user_id, event_name, resource_type, resource_id, ip_address, user_agent, new_state, timestamp)
					VALUES (gen_random_uuid(), $1, $2, 'API_ENDPOINT', $3, $4, $5, $6, CURRENT_TIMESTAMP)
				`
				_, err := db.Pool.Exec(ctx, query, entry.ActorID, entry.EventName, entry.Path, entry.IP, entry.UserAgent, entry.NewState)
				if err != nil {
					log.Printf("Gagal mencatat audit log: %v", err)
				}
				cancel()
			}
		}()
	})
}

// Menutup worker audit secara aman dan memastikan seluruh jejak audit tersimpan sebelum server mati.
func CloseAuditWorker() {
	auditLock.Lock()
	if auditClosed || auditQueue == nil {
		auditLock.Unlock()
		return
	}
	auditClosed = true
	close(auditQueue)
	auditLock.Unlock()
	auditWg.Wait()
}

// Mencatat jejak audit keamanan pada tabel database untuk setiap operasi mutasi data sensitif.
func AuditInterceptor(db *database.PostgresDB) gin.HandlerFunc {
	InitAuditWorker(db)

	return func(c *gin.Context) {
		c.Next()

		method := c.Request.Method
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
			path := c.Request.URL.Path
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
			statusCode := c.Writer.Status()

			stateData, _ := json.Marshal(map[string]interface{}{
				"status_code": statusCode,
				"method":      method,
				"path":        path,
			})

			auditLock.Lock()
			if !auditClosed && auditQueue != nil {
				select {
				case auditQueue <- auditEntry{
					ActorID:   actorID,
					EventName: eventName,
					Path:      path,
					IP:        ip,
					UserAgent: ua,
					NewState:  stateData,
				}:
				default:
					log.Printf("Peringatan: Antrean audit log penuh, mengabaikan pencatatan untuk %s", eventName)
				}
			}
			auditLock.Unlock()
		}
	}
}
