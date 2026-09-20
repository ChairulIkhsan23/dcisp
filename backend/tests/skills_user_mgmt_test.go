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

// Menguji katalog matriks keahlian teknis dan penugasan tingkat profisiensi keahlian pengguna (FR-006, T-031).
func TestSkillMatrixAndProficiencyAssignment(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := people.NewRepository(db)
	service := people.NewService(repo, db, nil, nil)
	ctx := context.Background()

	// 1. Buat master skill baru
	skillName := fmt.Sprintf("Golang Microservices %d", time.Now().UnixNano()%1000000)
	skill, err := service.CreateSkill(ctx, &people.CreateSkillRequest{
		Name:     skillName,
		Category: "BACKEND",
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, skill.ID)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM skills WHERE id = $1", skill.ID)
	})

	// 2. Setup test user
	testUserID := uuid.New()
	testEmail := fmt.Sprintf("skill_user_%s@dcisp.internal", testUserID.String()[:8])
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Skill User Tester', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_skills WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	// 3. Uji penolakan tingkat profisiensi di luar batas 1 sampai 5
	_, err = service.AssignUserSkill(ctx, testUserID, &people.AssignUserSkillRequest{
		SkillID:          skill.ID,
		ProficiencyLevel: 6, // Tidak valid
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "antara 1 sampai 5")

	// 4. Penugasan skill valid dengan tingkat kemahiran 4 (Mahir)
	us, err := service.AssignUserSkill(ctx, testUserID, &people.AssignUserSkillRequest{
		SkillID:          skill.ID,
		ProficiencyLevel: 4,
	})
	require.NoError(t, err)
	assert.Equal(t, 4, us.ProficiencyLevel)

	// 5. Ambil daftar keahlian pengguna
	userSkills, err := service.GetUserSkills(ctx, testUserID)
	require.NoError(t, err)
	assert.Len(t, userSkills, 1)
	assert.Equal(t, skillName, userSkills[0].SkillName)

	// 6. Pembaruan tingkat profisiensi (Upsert)
	usUpdated, err := service.AssignUserSkill(ctx, testUserID, &people.AssignUserSkillRequest{
		SkillID:          skill.ID,
		ProficiencyLevel: 5, // Naik ke tingkat 5
	})
	require.NoError(t, err)
	assert.Equal(t, 5, usUpdated.ProficiencyLevel)

	// 7. Hapus keahlian dari pengguna
	err = service.DeleteUserSkill(ctx, testUserID, skill.ID)
	require.NoError(t, err)

	userSkillsAfter, err := service.GetUserSkills(ctx, testUserID)
	require.NoError(t, err)
	assert.Len(t, userSkillsAfter, 0)
}

// Menguji pencarian peserta magang, penugasan pembimbing, dan plotting alokasi batch (T-032).
func TestUserManagementSearchMentorAndPlotting(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := people.NewRepository(db)
	service := people.NewService(repo, db, nil, nil)
	ctx := context.Background()

	// 1. Buat batch asal dan batch tujuan
	batch1, _, err := service.CreateBatch(ctx, &people.CreateBatchRequest{
		BatchCode: fmt.Sprintf("BATCH-P1-%d", time.Now().UnixNano()%1000000),
		Name:      "Batch Asal 2026",
		StartDate: "2026-09-01",
		EndDate:   "2027-02-28",
		Quota:     5,
		Status:    people.BatchStatusActive,
	})
	require.NoError(t, err)

	batch2, _, err := service.CreateBatch(ctx, &people.CreateBatchRequest{
		BatchCode: fmt.Sprintf("BATCH-P2-%d", time.Now().UnixNano()%1000000),
		Name:      "Batch Tujuan 2026",
		StartDate: "2026-09-01",
		EndDate:   "2027-02-28",
		Quota:     5,
		Status:    people.BatchStatusActive,
	})
	require.NoError(t, err)

	// 2. Setup user mentor dan user peserta
	mentorUserID := uuid.New()
	mentorEmail := fmt.Sprintf("mentor_%s@dcisp.internal", mentorUserID.String()[:8])
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Master Mentor Citra', 'ACTIVE')
	`, mentorUserID, mentorEmail)
	require.NoError(t, err)

	internUserID := uuid.New()
	internEmail := fmt.Sprintf("intern_search_%s@dcisp.internal", internUserID.String()[:8])
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Andi Searchable Adventurer', 'ACTIVE')
	`, internUserID, internEmail)
	require.NoError(t, err)

	intern, err := service.RegisterIntern(ctx, &people.RegisterInternRequest{
		UserID:   internUserID,
		BatchID:  batch1.ID,
		JoinDate: "2026-09-01",
		EndDate:  "2027-02-28",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM interns WHERE id = $1", intern.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id IN ($1, $2)", internUserID, mentorUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batch_funds WHERE batch_id IN ($1, $2)", batch1.ID, batch2.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id IN ($1, $2)", batch1.ID, batch2.ID)
	})

	// 3. Uji pencarian peserta magang (Search by Name / Keyword)
	searchRes, total, err := service.SearchInterns(ctx, "Andi Searchable", nil, nil, "", 1, 10)
	require.NoError(t, err)
	assert.True(t, total >= 1)
	assert.Equal(t, intern.ID, searchRes[0].ID)

	// 4. Uji penugasan pembimbing (Mentor Assignment)
	assignedIntern, err := service.AssignMentor(ctx, intern.ID, mentorUserID)
	require.NoError(t, err)
	assert.NotNil(t, assignedIntern.MentorID)
	assert.Equal(t, mentorUserID, *assignedIntern.MentorID)

	// 5. Uji plotting alokasi batch (Batch Plotting)
	plottedIntern, err := service.PlotBatch(ctx, intern.ID, batch2.ID)
	require.NoError(t, err)
	assert.Equal(t, batch2.ID, plottedIntern.BatchID)
	assert.Equal(t, "Batch Tujuan 2026", plottedIntern.BatchName)
}
