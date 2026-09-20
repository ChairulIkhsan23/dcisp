package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/people"
	"dcisp/backend/internal/modules/performance"
	"dcisp/backend/internal/shared/eventbus"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji mesin aturan XP tiga jalur terisolasi, buku besar append-only, batas saldo non-negatif, dan idempotensi (FR-025, BR-004, BRULE-PRF-001, BRULE-PRF-005, T-054).
func TestXPRulesEngineAndThreeSchemeIsolation(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := performance.NewRepository(db)
	service := performance.NewService(repo, db, nil, nil)
	ctx := context.Background()

	testUserID := uuid.New()
	testEmail := fmt.Sprintf("xp_hero_%s@dcisp.internal", testUserID.String()[:8])
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'XP Hero Tester', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	// Inisialisasi intern profil dengan rank awal Novice (min_xp = 0)
	rankNoviceID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	batchID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO batches (id, batch_code, name, start_date, end_date, quota, status)
		VALUES ($1, $2, 'Batch XP Test', '2026-01-01', '2026-12-31', 50, 'ACTIVE')
	`, batchID, fmt.Sprintf("BATCH-XP-%s", batchID.String()[:6]))
	require.NoError(t, err)

	internID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO interns (id, user_id, batch_id, current_rank_id, internship_xp, status, join_date, end_date)
		VALUES ($1, $2, $3, $4, 0, 'ACTIVE', '2026-01-01', '2026-12-31')
	`, internID, testUserID, batchID, rankNoviceID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM interns WHERE id = $1", internID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id = $1", batchID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	// 1. Skema 1: Internship XP (Penambahan +10 XP dari Check-In tepat waktu)
	ref1 := fmt.Sprintf("test_checkin_%s_%d", testUserID.String()[:8], time.Now().UnixNano())
	tx1, err := service.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         testUserID,
		EventTrigger:   "CHECK_IN_ON_TIME",
		ReferenceEvent: ref1,
	})
	require.NoError(t, err)
	assert.Equal(t, performance.XPSchemeInternship, tx1.Scheme)
	assert.Equal(t, 10, tx1.Points)
	assert.Equal(t, 10, tx1.RunningBalance)

	// 2. Uji Idempotensi: Memproses event dengan referensi yang sama persis tidak menambah mutasi baru
	txDuplicate, err := service.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         testUserID,
		EventTrigger:   "CHECK_IN_ON_TIME",
		ReferenceEvent: ref1,
	})
	require.NoError(t, err)
	assert.Equal(t, 10, txDuplicate.RunningBalance)

	// 3. Skema 2: Project XP (+30 XP dari tugas selesai tingkat menengah)
	ref2 := fmt.Sprintf("test_task_%s_%d", testUserID.String()[:8], time.Now().UnixNano())
	tx2, err := service.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         testUserID,
		EventTrigger:   "TASK_COMPLETED_MED",
		ReferenceEvent: ref2,
	})
	require.NoError(t, err)
	assert.Equal(t, performance.XPSchemeProject, tx2.Scheme)
	assert.Equal(t, 30, tx2.Points)
	assert.Equal(t, 30, tx2.RunningBalance)

	// 4. Skema 3: Alumni Contribution (+30 XP)
	ref3 := fmt.Sprintf("test_alumni_%s_%d", testUserID.String()[:8], time.Now().UnixNano())
	tx3, err := service.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         testUserID,
		EventTrigger:   "ALUMNI_TASK_COMPLETED",
		ReferenceEvent: ref3,
	})
	require.NoError(t, err)
	assert.Equal(t, performance.XPSchemeAlumni, tx3.Scheme)
	assert.Equal(t, 30, tx3.Points)
	assert.Equal(t, 30, tx3.RunningBalance)

	// 5. Verifikasi Isolasi 3 Jalur: Saldo ketiga skema tidak saling mencemari
	balances, err := service.GetXPBalance(ctx, testUserID)
	require.NoError(t, err)
	assert.Equal(t, 10, balances.InternshipXP)
	assert.Equal(t, 30, balances.ProjectXP)
	assert.Equal(t, 30, balances.AlumniContributionXP)

	// 6. Uji Batas Bawah Non-Negatif (Floor limit = 0 XP pada Internship XP - BRULE-PRF-005)
	refPenalty := fmt.Sprintf("test_penalty_%s_%d", testUserID.String()[:8], time.Now().UnixNano())
	penPoints := -50 // Penalti besar melebihi saldo 10
	txPen, err := service.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         testUserID,
		EventTrigger:   "ABSENT_UNAUTHORIZED",
		PointsOverride: &penPoints,
		ReferenceEvent: refPenalty,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, txPen.RunningBalance) // Tidak boleh negatif
}

// Menguji progresi kenaikan tingkat rank dan penegakan aturan tidak adanya penurunan otomatis (FR-026, BRULE-PRF-003, T-056).
func TestRankProgressionAndNoAutoDemotion(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := performance.NewRepository(db)
	service := performance.NewService(repo, db, nil, nil)
	ctx := context.Background()

	testUserID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Rank Hero Tester', 'ACTIVE')
	`, testUserID, fmt.Sprintf("rank_hero_%s@dcisp.internal", testUserID.String()[:8]))
	require.NoError(t, err)

	rankNoviceID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	batchID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO batches (id, batch_code, name, start_date, end_date, quota, status)
		VALUES ($1, $2, 'Batch Rank Test', '2026-01-01', '2026-12-31', 50, 'ACTIVE')
	`, batchID, fmt.Sprintf("BATCH-RK-%s", batchID.String()[:6]))
	require.NoError(t, err)

	internID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO interns (id, user_id, batch_id, current_rank_id, internship_xp, status, join_date, end_date)
		VALUES ($1, $2, $3, $4, 0, 'ACTIVE', '2026-01-01', '2026-12-31')
	`, internID, testUserID, batchID, rankNoviceID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM interns WHERE id = $1", internID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id = $1", batchID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	// 1. Naikkan XP melampaui ambang Tier 2: Apprentice (min_xp = 250)
	pts260 := 260
	refUp := fmt.Sprintf("rank_up_%s_%d", testUserID.String()[:8], time.Now().UnixNano())
	_, err = service.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         testUserID,
		EventTrigger:   "CHECK_IN_ON_TIME",
		PointsOverride: &pts260,
		ReferenceEvent: refUp,
	})
	require.NoError(t, err)

	// Verifikasi rank naik ke Tier 2: Apprentice
	var currentRankID uuid.UUID
	err = db.Pool.QueryRow(ctx, "SELECT current_rank_id FROM interns WHERE user_id = $1", testUserID).Scan(&currentRankID)
	require.NoError(t, err)
	rankApprenticeID := uuid.MustParse("30000000-0000-0000-0000-000000000002")
	assert.Equal(t, rankApprenticeID, currentRankID)

	// 2. Kenakan penalti pengurangan -50 XP (saldo menjadi 210 XP, di bawah ambang 250)
	ptsPenalty := -50
	refDown := fmt.Sprintf("rank_down_%s_%d", testUserID.String()[:8], time.Now().UnixNano())
	_, err = service.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         testUserID,
		EventTrigger:   "LATE_TIER_3",
		PointsOverride: &ptsPenalty,
		ReferenceEvent: refDown,
	})
	require.NoError(t, err)

	// Verifikasi aturan No Auto-Demotion (BRULE-PRF-003): Tingkat rank tetap Tier 2: Apprentice meskipun saldo 210 XP!
	var rankAfterPenalty uuid.UUID
	err = db.Pool.QueryRow(ctx, "SELECT current_rank_id FROM interns WHERE user_id = $1", testUserID).Scan(&rankAfterPenalty)
	require.NoError(t, err)
	assert.Equal(t, rankApprenticeID, rankAfterPenalty)
}

// Menguji evaluasi kinerja formal supervisor berbasis rubrik 0–100 dan penentuan tepat satu Top Performer per batch (FR-027, FR-028, BR-017, BR-018, T-058, T-059).
func TestSupervisorPerformanceEvaluationAndTopPerformer(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := performance.NewRepository(db)
	service := performance.NewService(repo, db, nil, nil)
	ctx := context.Background()

	// Setup Supervisor, Batch, dan 2 Peserta (A dan B)
	supervisorID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()
	batchID := uuid.New()

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Supervisor Guild Master', 'ACTIVE'),
		       ($3, $4, 'hash_pass', 'Intern Andi Top', 'ACTIVE'),
		       ($5, $6, 'hash_pass', 'Intern Budi Candidate', 'ACTIVE')
	`, supervisorID, fmt.Sprintf("spv_eval_%s@dcisp.internal", supervisorID.String()[:8]),
		userA, fmt.Sprintf("ua_eval_%s@dcisp.internal", userA.String()[:8]),
		userB, fmt.Sprintf("ub_eval_%s@dcisp.internal", userB.String()[:8]))
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO batches (id, batch_code, name, start_date, end_date, quota, status)
		VALUES ($1, $2, 'Batch Evaluasi Performa', '2026-01-01', '2026-12-31', 20, 'ACTIVE')
	`, batchID, fmt.Sprintf("BATCH-EVAL-%s", batchID.String()[:6]))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM performance_evaluations WHERE batch_id = $1", batchID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id = $1", batchID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id IN ($1, $2, $3)", supervisorID, userA, userB)
	})

	// 1. Evaluasi Peserta A: Skor Sangat Tinggi (Attendance: 95, Task: 90, Quality: 95, Rubric: 100 -> Komposit: 95.00)
	evalA, err := service.CreateEvaluation(ctx, supervisorID, &performance.CreateEvaluationRequest{
		UserID:                userA,
		BatchID:               batchID,
		PeriodType:            performance.PeriodEndOfBatch,
		AttendanceScore:       95.00,
		TaskDeliveryScore:     90.00,
		WorkQualityScore:      95.00,
		SupervisorRubricScore: 100.00,
		Notes:                 func(s string) *string { return &s }("Peserta teladan dengan dedikasi tinggi"),
	})
	require.NoError(t, err)
	assert.Equal(t, 95.00, evalA.CompositePerformanceScore)

	// 2. Evaluasi Peserta B: Skor Menengah (Attendance: 80, Task: 80, Quality: 80, Rubric: 80 -> Komposit: 80.00)
	evalB, err := service.CreateEvaluation(ctx, supervisorID, &performance.CreateEvaluationRequest{
		UserID:                userB,
		BatchID:               batchID,
		PeriodType:            performance.PeriodEndOfBatch,
		AttendanceScore:       80.00,
		TaskDeliveryScore:     80.00,
		WorkQualityScore:      80.00,
		SupervisorRubricScore: 80.00,
	})
	require.NoError(t, err)
	assert.Equal(t, 80.00, evalB.CompositePerformanceScore)

	// 3. Penentuan Top Performer Tunggal (Singular Top Performer per Batch - BRULE-PRF-004)
	topPerformer, err := service.DetermineTopPerformer(ctx, batchID, performance.PeriodEndOfBatch)
	require.NoError(t, err)
	assert.Equal(t, userA, topPerformer.WinnerUserID)
	assert.Equal(t, 95.00, topPerformer.CompositePerformanceScore)

	// Verifikasi di database bahwa hanya tepat 1 orang yang memiliki is_top_performer = true
	var topCount int
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM performance_evaluations
		WHERE batch_id = $1 AND period_type = $2 AND is_top_performer = TRUE
	`, batchID, performance.PeriodEndOfBatch).Scan(&topCount)
	require.NoError(t, err)
	assert.Equal(t, 1, topCount)

	_ = evalB
}

// Menguji pembukaan lencana prestasi pengguna secara idempoten dan matriks pertumbuhan keahlian (FR-029, FR-030, T-060, T-061).
func TestAchievementsAndSkillGrowthMatrix(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	peopleRepo := people.NewRepository(db)
	repo := performance.NewRepository(db)
	service := performance.NewService(repo, db, nil, nil)
	ctx := context.Background()

	testUserID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Achievement Skill Tester', 'ACTIVE')
	`, testUserID, fmt.Sprintf("ach_skill_%s@dcisp.internal", testUserID.String()[:8]))
	require.NoError(t, err)

	// 1. Buat master achievement
	achCode := fmt.Sprintf("EARLY_BIRD_%d", time.Now().UnixNano()%100000)
	ach, err := service.CreateAchievement(ctx, &performance.CreateAchievementRequest{
		Code:        achCode,
		Title:       "Early Bird Champion",
		Description: "Melakukan check-in presensi tepat waktu 5 hari berturut-turut",
		RewardXP:    50,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_achievements WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM achievements WHERE id = $1", ach.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_skills WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	// 2. Buka achievement pertama kali (Harus sukses dan is_new = true)
	isNew1, err := service.UnlockAchievement(ctx, testUserID, achCode)
	require.NoError(t, err)
	assert.True(t, isNew1)

	// 3. Buka achievement kedua kali (Idempoten: is_new = false)
	isNew2, err := service.UnlockAchievement(ctx, testUserID, achCode)
	require.NoError(t, err)
	assert.False(t, isNew2)

	// 4. Verifikasi daftar lencana prestasi pengguna
	myAchs, err := service.GetUserAchievements(ctx, testUserID)
	require.NoError(t, err)
	assert.Len(t, myAchs, 1)
	assert.Equal(t, "Early Bird Champion", myAchs[0].Title)

	// 5. Uji Agregasi Skill Growth Matrix (FR-030, T-061)
	skillID := uuid.New()
	s := &people.Skill{ID: skillID, Name: fmt.Sprintf("Go Architecture %d", time.Now().UnixNano()%100000), Category: "BACKEND"}
	require.NoError(t, peopleRepo.CreateSkill(ctx, s))
	require.NoError(t, peopleRepo.UpsertUserSkill(ctx, &people.UserSkill{ID: uuid.New(), UserID: testUserID, SkillID: skillID, ProficiencyLevel: 4}))

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM skills WHERE id = $1", skillID)
	})

	matrix, err := service.GetSkillGrowthMatrix(ctx, testUserID)
	require.NoError(t, err)
	assert.NotNil(t, matrix)
	assert.Len(t, matrix.Skills, 1)
	assert.Equal(t, s.Name, matrix.Skills[0].SkillName)
	assert.Equal(t, 4, matrix.Skills[0].ProficiencyLevel)
}

// Menguji integrasi asinkron pemrosesan event kehadiran dan penyelesaian tugas ke mutasi XP (T-055, T-057).
func TestAsyncAttendanceAndTaskXPIntegration(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	bus := eventbus.NewEventBus(nil)
	repo := performance.NewRepository(db)
	service := performance.NewService(repo, db, nil, bus)
	ctx := context.Background()

	testUserID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Async XP Tester', 'ACTIVE')
	`, testUserID, fmt.Sprintf("async_xp_%s@dcisp.internal", testUserID.String()[:8]))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	// 1. Publikasikan event presensi tepat waktu (attendance.scanned)
	tsNow := time.Now().UTC().Format(time.RFC3339)
	err = bus.Publish(ctx, eventbus.DomainEvent{
		Type:        "attendance.scanned",
		AggregateID: testUserID.String(),
		Payload: map[string]interface{}{
			"user_id":    testUserID.String(),
			"event_type": "CHECK_IN",
			"is_late":    false,
			"timestamp":  tsNow,
		},
	})
	require.NoError(t, err)

	// Beri jeda sejenak untuk dispatch asinkron
	time.Sleep(100 * time.Millisecond)

	// 2. Publikasikan event penyelesaian tugas proyek berbobot tinggi (task.report_reviewed)
	taskID := uuid.New()
	err = bus.Publish(ctx, eventbus.DomainEvent{
		Type:        "task.report_reviewed",
		AggregateID: taskID.String(),
		Payload: map[string]interface{}{
			"status":            "APPROVED",
			"task_id":           taskID.String(),
			"assignee_id":       testUserID.String(),
			"difficulty_weight": float64(4), // High difficulty -> +50 Project XP
		},
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// 3. Verifikasi saldo XP terakumulasi pada kedua skema
	balances, err := service.GetXPBalance(ctx, testUserID)
	require.NoError(t, err)
	assert.Equal(t, 10, balances.InternshipXP) // Dari Check-In tepat waktu
	assert.Equal(t, 50, balances.ProjectXP)    // Dari Task Completed High
}
