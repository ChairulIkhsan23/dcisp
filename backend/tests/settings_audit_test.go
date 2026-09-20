package tests

import (
	"context"
	"testing"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/system"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji siklus CRUD pengaturan sistem terpusat dan mekanisme masking nilai sensitif terenkripsi.
func TestSystemSettingsCRUDAndMasking(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	repo := system.NewSettingsRepository(db)
	service := system.NewSettingsService(repo, rdb)
	ctx := context.Background()

	testKey := "test.platform.maintenance_mode"
	t.Cleanup(func() {
		_ = service.DeleteSetting(context.Background(), testKey)
	})

	// 1. Simpan pengaturan baru
	req := &system.SetSettingRequest{
		SettingKey:   testKey,
		SettingValue: "true",
		IsEncrypted:  false,
		Description:  "Status mode pemeliharaan platform",
	}
	s, err := service.SetSetting(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, testKey, s.SettingKey)
	assert.Equal(t, "true", s.SettingValue)

	// 2. Ambil pengaturan (memverifikasi pembacaan dari DB / Redis)
	fetched, err := service.GetSetting(ctx, testKey)
	require.NoError(t, err)
	assert.Equal(t, "true", fetched.SettingValue)

	// 3. Simpan pengaturan terenkripsi dan verifikasi penyamaran nilai pada List
	secretKey := "test.secret.webhook_token"
	t.Cleanup(func() {
		_ = service.DeleteSetting(context.Background(), secretKey)
	})

	_, err = service.SetSetting(ctx, &system.SetSettingRequest{
		SettingKey:   secretKey,
		SettingValue: "super_secret_token_value",
		IsEncrypted:  true,
		Description:  "Token rahasia webhook",
	})
	require.NoError(t, err)

	allSettings, err := service.ListSettings(ctx)
	require.NoError(t, err)
	foundSecret := false
	for _, setting := range allSettings {
		if setting.SettingKey == secretKey {
			foundSecret = true
			assert.Equal(t, "********", setting.SettingValue) // Harus tersamarkan
		}
	}
	assert.True(t, foundSecret)
}

// Menguji pencatatan mutasi audit terstruktur dan penyaringan otomatis data kredensial sensitif.
func TestStructuredAuditLoggingAndRedaction(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	service := system.NewAuditService(db)
	ctx := context.Background()

	testUserID := uuid.New()
	testEmail := "audit_target_" + testUserID.String()[:8] + "@dcisp.internal"

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash123', 'Audit Target User', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	testResourceID := "policy_" + uuid.New().String()[:8]

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM audit_logs WHERE resource_id = $1", testResourceID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	// 1. Verifikasi penyaringan data sensitif pada helper RedactSensitiveData
	rawPayload := map[string]interface{}{
		"email":        "user@dcisp.internal",
		"password":     "SecretPassword123!",
		"jwt_secret":   "my_jwt_secret",
		"access_token": "bearer_token_string",
		"normal_field": "safe_value",
		"nested_object": map[string]interface{}{
			"secret_key": "nested_secret",
			"score":      100,
		},
	}
	redacted := system.RedactSensitiveData(rawPayload)
	assert.Equal(t, "[REDACTED]", redacted["password"])
	assert.Equal(t, "[REDACTED]", redacted["jwt_secret"])
	assert.Equal(t, "[REDACTED]", redacted["access_token"])
	assert.Equal(t, "safe_value", redacted["normal_field"])

	nested := redacted["nested_object"].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", nested["secret_key"])
	assert.Equal(t, 100, nested["score"])

	// 2. Simpan rekaman mutasi audit ke database
	err = service.RecordMutation(ctx, system.MutationAuditEntry{
		UserID:       &testUserID,
		EventName:    "UPDATE_SECURITY_POLICY",
		ResourceType: "POLICY",
		ResourceID:   testResourceID,
		IPAddress:    "127.0.0.1",
		UserAgent:    "DCISP-Automated-Test/1.0",
		OldState:     rawPayload,
		NewState: map[string]interface{}{
			"status": "APPROVED",
		},
	})
	require.NoError(t, err)

	// 3. Verifikasi pembacaan jejak audit dari database
	logs, total, err := service.ListAuditLogs(ctx, "POLICY", testResourceID, &testUserID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, logs, 1)

	loggedRecord := logs[0]
	assert.Equal(t, "UPDATE_SECURITY_POLICY", loggedRecord.EventName)
	assert.Equal(t, testResourceID, loggedRecord.ResourceID)
	assert.NotContains(t, string(loggedRecord.OldState), "SecretPassword123!")
	assert.Contains(t, string(loggedRecord.OldState), "[REDACTED]")
}
