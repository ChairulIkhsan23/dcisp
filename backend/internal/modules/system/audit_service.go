package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuditLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       *uuid.UUID      `json:"user_id" db:"user_id"`
	EventName    string          `json:"event_name" db:"event_name"`
	ResourceType string          `json:"resource_type" db:"resource_type"`
	ResourceID   string          `json:"resource_id" db:"resource_id"`
	IPAddress    string          `json:"ip_address" db:"ip_address"`
	UserAgent    string          `json:"user_agent" db:"user_agent"`
	OldState     json.RawMessage `json:"old_state" db:"old_state"`
	NewState     json.RawMessage `json:"new_state" db:"new_state"`
	Timestamp    time.Time       `json:"timestamp" db:"timestamp"`
}

type MutationAuditEntry struct {
	UserID       *uuid.UUID             `json:"user_id"`
	EventName    string                 `json:"event_name"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	OldState     map[string]interface{} `json:"old_state"`
	NewState     map[string]interface{} `json:"new_state"`
}

type AuditService struct {
	db *database.PostgresDB
}

// Menginisialisasi instance baru service audit logging terpusat.
func NewAuditService(db *database.PostgresDB) *AuditService {
	return &AuditService{db: db}
}

// Menyaring data sensitif seperti kata sandi, token, dan kunci rahasia dari rekaman audit.
func RedactSensitiveData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}
	redacted := make(map[string]interface{})
	sensitiveKeys := map[string]bool{
		"password":             true,
		"password_hash":        true,
		"token":                true,
		"access_token":         true,
		"refresh_token":        true,
		"jwt_secret":           true,
		"secret":               true,
		"secret_key":           true,
		"api_key":              true,
		"r2_secret_access_key": true,
		"db_password":          true,
		"redis_password":       true,
	}

	for k, v := range data {
		lowerKey := strings.ToLower(k)
		if sensitiveKeys[lowerKey] || strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "secret") {
			redacted[k] = "[REDACTED]"
		} else if nestedMap, ok := v.(map[string]interface{}); ok {
			redacted[k] = RedactSensitiveData(nestedMap)
		} else {
			redacted[k] = v
		}
	}
	return redacted
}

// Mencatat peristiwa mutasi data secara permanen dan aman ke tabel audit_logs.
func (s *AuditService) RecordMutation(ctx context.Context, entry MutationAuditEntry) error {
	redactedOld := RedactSensitiveData(entry.OldState)
	redactedNew := RedactSensitiveData(entry.NewState)

	oldBytes, err := json.Marshal(redactedOld)
	if err != nil {
		oldBytes = []byte("{}")
	}

	newBytes, err := json.Marshal(redactedNew)
	if err != nil {
		newBytes = []byte("{}")
	}

	query := `
		INSERT INTO audit_logs (id, user_id, event_name, resource_type, resource_id, ip_address, user_agent, old_state, new_state, timestamp)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
	`
	_, err = s.db.Pool.Exec(ctx, query, entry.UserID, entry.EventName, entry.ResourceType, entry.ResourceID, entry.IPAddress, entry.UserAgent, oldBytes, newBytes)
	if err != nil {
		return fmt.Errorf("gagal mencatat mutasi audit: %w", err)
	}
	return nil
}

// Mengambil daftar jejak log audit dengan filter dan paginasi untuk kebutuhan investigasi kepatuhan.
func (s *AuditService) ListAuditLogs(ctx context.Context, resourceType, resourceID string, userID *uuid.UUID, page, limit int) ([]AuditLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	baseQuery := `FROM audit_logs WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if resourceType != "" {
		baseQuery += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, resourceType)
		argIdx++
	}

	if resourceID != "" {
		baseQuery += fmt.Sprintf(" AND resource_id = $%d", argIdx)
		args = append(args, resourceID)
		argIdx++
	}

	if userID != nil {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *userID)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total log audit: %w", err)
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, user_id, event_name, resource_type, resource_id, ip_address, user_agent, old_state, new_state, timestamp
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar log audit: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.EventName, &l.ResourceType, &l.ResourceID, &l.IPAddress, &l.UserAgent, &l.OldState, &l.NewState, &l.Timestamp); err == nil {
			logs = append(logs, l)
		}
	}

	return logs, total, nil
}

// Mengambil satu rekaman jejak audit berdasarkan identitas unik.
func (s *AuditService) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*AuditLog, error) {
	query := `
		SELECT id, user_id, event_name, resource_type, resource_id, ip_address, user_agent, old_state, new_state, timestamp
		FROM audit_logs
		WHERE id = $1
	`
	var l AuditLog
	err := s.db.Pool.QueryRow(ctx, query, id).Scan(&l.ID, &l.UserID, &l.EventName, &l.ResourceType, &l.ResourceID, &l.IPAddress, &l.UserAgent, &l.OldState, &l.NewState, &l.Timestamp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari log audit berdasarkan id: %w", err)
	}
	return &l, nil
}
