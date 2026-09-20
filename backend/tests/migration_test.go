package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji eksekusi migrasi penambahan dan penghapusan indeks trigram GIN pada database PostgreSQL.
func TestTrigramIndexesMigration(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	upPath := filepath.Join("..", "migrations", "000003_add_trgm_indexes.up.sql")
	upSQL, err := os.ReadFile(upPath)
	require.NoError(t, err)

	downPath := filepath.Join("..", "migrations", "000003_add_trgm_indexes.down.sql")
	downSQL, err := os.ReadFile(downPath)
	require.NoError(t, err)

	// 1. Eksekusi Migrasi UP
	_, err = db.Pool.Exec(ctx, string(upSQL))
	require.NoError(t, err)

	// Verifikasi indeks ada di pg_indexes
	var count int
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE indexname IN ('idx_users_name_trgm', 'idx_projects_title_trgm')
	`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// 2. Eksekusi Migrasi DOWN
	_, err = db.Pool.Exec(ctx, string(downSQL))
	require.NoError(t, err)

	// Verifikasi indeks terhapus
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE indexname IN ('idx_users_name_trgm', 'idx_projects_title_trgm')
	`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// 3. Terapkan kembali UP agar database memiliki indeks aktif
	_, err = db.Pool.Exec(ctx, string(upSQL))
	require.NoError(t, err)
}
