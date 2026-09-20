package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/people"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji seluruh alur mesin status siklus hidup peserta magang, penolakan transisi ilegal, dan transisi kelulusan ke alumni (FR-002, FR-003, BR-003).
func TestInternLifecycleStateMachineAndAlumniTransition(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := people.NewRepository(db)
	service := people.NewService(repo, db, nil, nil)
	ctx := context.Background()

	// 1. Setup User akun dan Batch
	testUserID := uuid.New()
	testEmail := fmt.Sprintf("hero_%s@dcisp.internal", testUserID.String()[:8])
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Intern Lifecycle Hero', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	// Petakan peran awal ke INTERN
	internRoleID := uuid.MustParse("10000000-0000-0000-0000-000000000009")
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO user_roles (id, user_id, role_id, scope_id)
		VALUES (gen_random_uuid(), $1, $2, (SELECT id FROM scopes WHERE scope_type = 'OWN_DATA' LIMIT 1))
	`, testUserID, internRoleID)
	require.NoError(t, err)

	batch, _, err := service.CreateBatch(ctx, &people.CreateBatchRequest{
		BatchCode: fmt.Sprintf("BATCH-LC-%d", time.Now().UnixNano()%1000000),
		Name:      "Batch Lifecycle Testing",
		StartDate: "2026-09-01",
		EndDate:   "2027-02-28",
		Quota:     10,
		Status:    people.BatchStatusActive,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM alumni WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM interns WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_roles WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batch_funds WHERE batch_id = $1", batch.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id = $1", batch.ID)
	})

	// 2. Registrasi peserta (Status awal: APPLICANT)
	intern, err := service.RegisterIntern(ctx, &people.RegisterInternRequest{
		UserID:   testUserID,
		BatchID:  batch.ID,
		JoinDate: "2026-09-01",
		EndDate:  "2027-02-28",
	})
	require.NoError(t, err)
	assert.Equal(t, people.InternStatusApplicant, intern.Status)

	// 3. Uji penolakan transisi ilegal: APPLICANT -> GRADUATED (Harus ditolak)
	_, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusGraduated, "Langsung lulus ilegal")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak diizinkan oleh mesin status")

	// 4. Uji penolakan transisi ke status yang sama
	_, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusApplicant, "Status sama")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sudah")

	// 5. Alur sah: APPLICANT -> ONBOARDING
	intern, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusOnboarding, "Diterima magang")
	require.NoError(t, err)
	assert.Equal(t, people.InternStatusOnboarding, intern.Status)

	// 6. Alur sah: ONBOARDING -> ACTIVE
	intern, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusActive, "Aktivasi program")
	require.NoError(t, err)
	assert.Equal(t, people.InternStatusActive, intern.Status)

	// 7. Alur sah: ACTIVE -> ON_LEAVE
	intern, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusOnLeave, "Cuti akademik")
	require.NoError(t, err)
	assert.Equal(t, people.InternStatusOnLeave, intern.Status)

	// 8. Uji penolakan transisi ilegal dari ON_LEAVE -> SUSPENDED
	_, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusSuspended, "Skorsing saat cuti")
	assert.Error(t, err)

	// 9. Alur sah: ON_LEAVE -> ACTIVE
	intern, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusActive, "Kembali aktif dari cuti")
	require.NoError(t, err)
	assert.Equal(t, people.InternStatusActive, intern.Status)

	// Simulasikan perolehan XP magang
	_, err = db.Pool.Exec(ctx, "UPDATE interns SET internship_xp = 500 WHERE id = $1", intern.ID)
	require.NoError(t, err)

	// 10. Alur sah kelulusan: ACTIVE -> GRADUATED (FR-003, BRULE-PEO-001)
	intern, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusGraduated, "Lulus magang dengan predikat memuaskan")
	require.NoError(t, err)
	assert.Equal(t, people.InternStatusGraduated, intern.Status)

	// 11. Verifikasi pembuatan profil Alumni secara otomatis
	alumni, err := service.GetAlumniByUserID(ctx, testUserID)
	require.NoError(t, err)
	assert.NotNil(t, alumni)
	assert.Equal(t, testUserID, alumni.UserID)
	assert.Equal(t, batch.ID, alumni.BatchID)
	assert.Equal(t, 0, alumni.AlumniXP) // Isolasi skema poin alumni (BRULE-PEO-002)

	// Verifikasi rekam jejak XP magang tetap utuh dan terisolasi
	internAfterGrad, err := service.GetInternByID(ctx, intern.ID)
	require.NoError(t, err)
	assert.Equal(t, 500, internAfterGrad.InternshipXP)

	// 12. Verifikasi pengalihan peran pengguna menjadi ALUMNI pada tabel user_roles
	var roleCount int
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.name = 'ALUMNI'
	`, testUserID).Scan(&roleCount)
	require.NoError(t, err)
	assert.Equal(t, 1, roleCount)

	// 13. Verifikasi akun pengguna tetap aktif (Permanent Account Retention - BR-003)
	var userStatus string
	err = db.Pool.QueryRow(ctx, "SELECT status FROM users WHERE id = $1", testUserID).Scan(&userStatus)
	require.NoError(t, err)
	assert.Equal(t, "ACTIVE", userStatus)

	// 14. Uji state akhir: GRADUATED tidak dapat bertransisi ke status lain lagi
	_, err = service.ChangeInternStatus(ctx, intern.ID, people.InternStatusActive, "Mencoba aktif lagi")
	assert.Error(t, err)
}
