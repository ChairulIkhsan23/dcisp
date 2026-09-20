package system

import (
	"time"

	"github.com/google/uuid"
)

type SystemSetting struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SettingKey   string    `json:"setting_key" db:"setting_key"`
	SettingValue string    `json:"setting_value" db:"setting_value"`
	IsEncrypted  bool      `json:"is_encrypted" db:"is_encrypted"`
	Description  string    `json:"description" db:"description"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type SetSettingRequest struct {
	SettingKey   string `json:"setting_key" binding:"required"`
	SettingValue string `json:"setting_value" binding:"required"`
	IsEncrypted  bool   `json:"is_encrypted"`
	Description  string `json:"description"`
}

type UpdateSettingRequest struct {
	SettingValue string `json:"setting_value" binding:"required"`
	IsEncrypted  *bool  `json:"is_encrypted"`
	Description  string `json:"description"`
}
