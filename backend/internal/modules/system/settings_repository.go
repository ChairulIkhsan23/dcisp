package system

import (
	"context"
	"errors"
	"fmt"

	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SettingsRepository struct {
	db *database.PostgresDB
}

// Menginisialisasi instance baru repository pengaturan sistem terpusat.
func NewSettingsRepository(db *database.PostgresDB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Menyimpan entitas pengaturan sistem baru ke dalam tabel system_settings.
func (r *SettingsRepository) Create(ctx context.Context, s *SystemSetting) error {
	query := `
		INSERT INTO system_settings (id, setting_key, setting_value, is_encrypted, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, s.ID, s.SettingKey, s.SettingValue, s.IsEncrypted, s.Description)
	if err != nil {
		return fmt.Errorf("gagal menyimpan pengaturan sistem: %w", err)
	}
	return nil
}

// Mengambil data pengaturan sistem dari database berdasarkan kunci unik.
func (r *SettingsRepository) GetByKey(ctx context.Context, key string) (*SystemSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, is_encrypted, description, created_at, updated_at
		FROM system_settings
		WHERE setting_key = $1
	`
	var s SystemSetting
	err := r.db.Pool.QueryRow(ctx, query, key).Scan(&s.ID, &s.SettingKey, &s.SettingValue, &s.IsEncrypted, &s.Description, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari pengaturan sistem: %w", err)
	}
	return &s, nil
}

// Mengambil seluruh daftar pengaturan sistem yang terdaftar di database.
func (r *SettingsRepository) List(ctx context.Context) ([]SystemSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, is_encrypted, description, created_at, updated_at
		FROM system_settings
		ORDER BY setting_key ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pengaturan sistem: %w", err)
	}
	defer rows.Close()

	var settings []SystemSetting
	for rows.Next() {
		var s SystemSetting
		if err := rows.Scan(&s.ID, &s.SettingKey, &s.SettingValue, &s.IsEncrypted, &s.Description, &s.CreatedAt, &s.UpdatedAt); err == nil {
			settings = append(settings, s)
		}
	}
	return settings, nil
}

// Memperbarui nilai dan konfigurasi pengaturan sistem yang sudah ada.
func (r *SettingsRepository) Update(ctx context.Context, s *SystemSetting) error {
	query := `
		UPDATE system_settings
		SET setting_value = $2, is_encrypted = $3, description = $4, updated_at = CURRENT_TIMESTAMP
		WHERE setting_key = $1
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, s.SettingKey, s.SettingValue, s.IsEncrypted, s.Description)
	if err != nil {
		return fmt.Errorf("gagal memperbarui pengaturan sistem: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("pengaturan sistem tidak ditemukan")
	}
	return nil
}

// Menghapus data pengaturan sistem berdasarkan kunci unik.
func (r *SettingsRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM system_settings WHERE setting_key = $1`
	cmdTag, err := r.db.Pool.Exec(ctx, query, key)
	if err != nil {
		return fmt.Errorf("gagal menghapus pengaturan sistem: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("pengaturan sistem tidak ditemukan")
	}
	return nil
}
