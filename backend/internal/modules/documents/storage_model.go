package documents

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FileMetadata struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	Disk      string          `json:"disk" db:"disk"`
	PathKey   string          `json:"path_key" db:"path_key"`
	Filename  string          `json:"filename" db:"filename"`
	MimeType  string          `json:"mime_type" db:"mime_type"`
	SizeBytes int64           `json:"size_bytes" db:"size_bytes"`
	Metadata  json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type PresignedUploadRequest struct {
	Folder    string `json:"folder" binding:"required"`
	Filename  string `json:"filename" binding:"required"`
	MimeType  string `json:"mime_type" binding:"required"`
	SizeBytes int64  `json:"size_bytes" binding:"required,min=1"`
}

type PresignedUploadResponse struct {
	FileID           uuid.UUID `json:"file_id"`
	UploadURL        string    `json:"upload_url"`
	PathKey          string    `json:"path_key"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
}
