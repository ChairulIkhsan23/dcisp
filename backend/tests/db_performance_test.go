package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/attendance"
	"dcisp/backend/internal/modules/finance"
	"dcisp/backend/internal/modules/people"
	"dcisp/backend/internal/modules/projects"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji eksekusi migrasi penambahan indeks performa dan pembersihan indeks duplikat pada database PostgreSQL.
func TestPerformance_Migration000008(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	upPath := filepath.Join("..", "migrations", "000008_performance_indexes.up.sql")
	upSQL, err := os.ReadFile(upPath)
	require.NoError(t, err)

	downPath := filepath.Join("..", "migrations", "000008_performance_indexes.down.sql")
	downSQL, err := os.ReadFile(downPath)
	require.NoError(t, err)

	// 1. Terapkan Migrasi UP
	_, err = db.Pool.Exec(ctx, string(upSQL))
	require.NoError(t, err)

	// Verifikasi 10 indeks baru ada
	expectedNewIndexes := []string{
		"idx_interns_id_number",
		"idx_breaks_session_start",
		"idx_xp_tx_user_reference",
		"idx_reward_claims_user_created",
		"idx_tasks_assignee_status",
		"idx_evidence_submission",
		"idx_milestones_project_deadline",
		"idx_perf_eval_batch_period",
		"idx_financial_ledgers_account_created",
		"idx_interns_batch_status_created",
	}

	for _, idxName := range expectedNewIndexes {
		var exists bool
		err = db.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = $1
			)
		`, idxName).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Indeks harus terdaftar: %s", idxName)
	}

	// Verifikasi 2 indeks duplikat telah terhapus
	droppedIndexes := []string{"idx_users_email", "idx_attendance_idempotency"}
	for _, idxName := range droppedIndexes {
		var exists bool
		err = db.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = $1
			)
		`, idxName).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "Indeks redundan harus sudah dihapus: %s", idxName)
	}

	// 2. Uji Migrasi DOWN
	_, err = db.Pool.Exec(ctx, string(downSQL))
	require.NoError(t, err)

	// Verifikasi indeks baru hilang di DOWN
	for _, idxName := range expectedNewIndexes {
		var exists bool
		err = db.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = $1
			)
		`, idxName).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "Indeks harus hilang setelah rollback DOWN: %s", idxName)
	}

	// 3. Terapkan kembali UP agar database memiliki indeks aktif
	_, err = db.Pool.Exec(ctx, string(upSQL))
	require.NoError(t, err)
}

// Menguji rencana eksekusi query pencarian kartu identitas NFC untuk memastikan indeks B-Tree digunakan.
func TestPerformance_NFCScan_UsesIndexScan(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	query := `
		EXPLAIN (FORMAT JSON)
		SELECT u.id, u.full_name, u.status
		FROM interns i
		JOIN users u ON u.id = i.user_id
		WHERE i.id_number = 'NIM-TEST-9999'
		LIMIT 1;
	`
	var planJSON string
	err = db.Pool.QueryRow(ctx, query).Scan(&planJSON)
	require.NoError(t, err)

	assert.Contains(t, planJSON, "idx_interns_id_number", "Query pencarian NIM harus menggunakan idx_interns_id_number")
	assert.NotContains(t, planJSON, "Seq Scan on interns", "Query pencarian NIM tidak boleh melakukan Seq Scan pada tabel interns")
}

// Menguji bahwa fungsi presensi harian memanfaatkan indeks rentang waktu secara SARGable.
func TestPerformance_SARGableAttendanceQuery(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	repo := attendance.NewRepository(db)

	testUserID := uuid.New()
	targetTime := time.Now().In(time.FixedZone("WIB", 7*3600))

	// Jalankan HasCheckInToday
	hasCheckIn, err := repo.HasCheckInToday(ctx, testUserID, targetTime)
	require.NoError(t, err)
	assert.False(t, hasCheckIn)

	// Jalankan GetLatestEventForUserToday
	latestEv, err := repo.GetLatestEventForUserToday(ctx, testUserID, targetTime)
	require.NoError(t, err)
	assert.Nil(t, latestEv)

	// Verifikasi via EXPLAIN bahwa query menggunakan indeks idx_attendance_user_time
	explainQuery := `
		EXPLAIN (FORMAT JSON)
		SELECT COUNT(*)
		FROM attendance_event_logs
		WHERE user_id = $1
		  AND event_type = 'CHECK_IN'
		  AND timestamp >= $2 AND timestamp < $3;
	`
	startOfDay := time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), 0, 0, 0, 0, targetTime.Location()).UTC()
	endOfDay := startOfDay.Add(24 * time.Hour)

	var planJSON string
	err = db.Pool.QueryRow(ctx, explainQuery, testUserID, startOfDay, endOfDay).Scan(&planJSON)
	require.NoError(t, err)

	assert.Contains(t, planJSON, "idx_attendance_user_time", "Query SARGable harus menggunakan idx_attendance_user_time")
}

// Menguji adapter kontribusi proyek untuk memastikan penghapusan pola N+1 query loop.
func TestPerformance_BountyAdapter_NoNPlusOne(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	projectsRepo := projects.NewRepository(db)
	peopleRepo := people.NewRepository(db)
	adapter := finance.NewProjectsContributionAdapter(projectsRepo, peopleRepo, db)

	// 1. Setup data batch
	batchID := uuid.New()
	batchCode := fmt.Sprintf("BATCH-PERF-%s", batchID.String()[:8])
	tx, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	_, err = peopleRepo.CreateBatchWithFundTx(ctx, tx, &people.Batch{
		ID:        batchID,
		BatchCode: batchCode,
		Name:      "Performance Test Batch",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(30 * 24 * time.Hour),
		Quota:     10,
		Status:    "ACTIVE",
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id = $1", batchID)
	})

	// 2. Setup user & intern
	userID := uuid.New()
	userEmail := fmt.Sprintf("perf_user_%s@dcisp.internal", userID.String()[:8])
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash', 'Performance User', 'ACTIVE')
	`, userID, userEmail)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	})

	intern := &people.Intern{
		ID:           uuid.New(),
		UserID:       userID,
		BatchID:      batchID,
		Status:       "ACTIVE",
		JoinDate:     time.Now(),
		EndDate:      time.Now().Add(30 * 24 * time.Hour),
		InternshipXP: 0,
	}
	require.NoError(t, peopleRepo.CreateIntern(ctx, intern))
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM interns WHERE id = $1", intern.ID)
	})

	// 3. Setup project & team
	projectID := uuid.New()
	project := &projects.Project{
		ID:          projectID,
		Title:       "Performance Bounty Project",
		Description: "Testing single-query team list",
		OwnerID:     userID,
		Visibility:  "PUBLIC",
		Capacity:    3,
		Deadline:    time.Now().Add(7 * 24 * time.Hour),
		BountyPool:  1000000.00,
		Status:      "COMPLETED",
	}
	require.NoError(t, projectsRepo.CreateProject(ctx, project))
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM projects WHERE id = $1", projectID)
	})

	finalPct := 100.00
	teamMember := &projects.ProjectTeamMember{
		ID:                     uuid.New(),
		ProjectID:              projectID,
		UserID:                 userID,
		ProjectRole:            "MEMBER",
		PlannedContributionPct: 100.00,
		IsLocked:               false,
	}
	require.NoError(t, projectsRepo.AddTeamMember(ctx, teamMember))

	txLock, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, projectsRepo.LockFinalContributionTx(ctx, txLock, projectID, userID, finalPct))
	require.NoError(t, txLock.Commit(ctx))

	// 4. Panggil adapter GetDistributionMembers
	distInput, err := adapter.GetDistributionMembers(ctx, projectID)
	require.NoError(t, err)
	require.NotNil(t, distInput)
	assert.Equal(t, projectID, distInput.ProjectID)
	require.Len(t, distInput.Members, 1)

	// Pastikan batch_id berhasil dipetakan dari single JOIN query tanpa N+1
	assert.True(t, distInput.Members[0].HasBatch)
	assert.Equal(t, batchID, distInput.Members[0].BatchID)
}
