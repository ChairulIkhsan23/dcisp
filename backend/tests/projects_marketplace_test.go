package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/people"
	"dcisp/backend/internal/modules/projects"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji bursa proyek, penyaringan visibilitas berdasarkan peran pengguna, dan isolasi akses proyek privat (FR-016, BR-003, T-045).
func TestProjectMarketplaceAndVisibilityRules(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := projects.NewRepository(db)
	service := projects.NewService(repo, db, nil, nil, nil)
	ctx := context.Background()

	// Helper penugasan peran dinamis (otorisasi berbasis data permissions, BR-001)
	assignTestRole := func(userID uuid.UUID, roleID, scopeID string) {
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id, scope_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, role_id, scope_id) DO NOTHING
		`, userID, roleID, scopeID)
		require.NoError(t, err)
	}
	const (
		rolePM      = "10000000-0000-0000-0000-000000000004"
		roleScanner = "10000000-0000-0000-0000-000000000008"
		roleIntern  = "10000000-0000-0000-0000-000000000009"
		roleAlumni  = "10000000-0000-0000-0000-000000000010"
		scopeProj   = "20000000-0000-0000-0000-000000000004"
		scopeOwn    = "20000000-0000-0000-0000-000000000007"
		scopePublic = "20000000-0000-0000-0000-000000000008"
		scopeScan   = "20000000-0000-0000-0000-000000000006"
	)

	createTestUser := func(prefix, fullName string) uuid.UUID {
		uID := uuid.New()
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, status)
			VALUES ($1, $2, 'hash_pass', $3, 'ACTIVE')
		`, uID, fmt.Sprintf("%s_%s@dcisp.internal", prefix, uID.String()[:8]), fullName)
		require.NoError(t, err)
		return uID
	}

	// Setup Project Owner (PROJECT_MANAGER)
	ownerID := createTestUser("pm", "Project Manager Dimas")
	assignTestRole(ownerID, rolePM, scopeProj)

	// Setup aktor matriks otorisasi (F-RBAC-02)
	alumniID := createTestUser("alumni", "Alumni Budi")
	assignTestRole(alumniID, roleAlumni, scopePublic)
	internID := createTestUser("intern", "Intern Andi")
	assignTestRole(internID, roleIntern, scopeOwn)
	scannerID := createTestUser("scanner", "Gatekeeper Reihan")
	assignTestRole(scannerID, roleScanner, scopeScan)
	memberID := createTestUser("member", "Member Candra")
	assignTestRole(memberID, roleIntern, scopeOwn)
	outsiderID := uuid.New() // tanpa peran sama sekali

	allTestUsers := []uuid.UUID{ownerID, alumniID, internID, scannerID, memberID}

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_teams WHERE user_id = ANY($1)", allTestUsers)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM projects WHERE owner_id = $1", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_roles WHERE user_id = ANY($1)", allTestUsers)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = ANY($1)", allTestUsers)
	})

	// 1. Buat 3 proyek dengan visibilitas berbeda
	pPublic, err := service.CreateProject(ctx, ownerID, &projects.CreateProjectRequest{
		Title:       "Proyek Terbuka Alumni & Magang",
		Description: "Pengembangan portal publik",
		Visibility:  projects.VisibilityPublic,
		Capacity:    5,
		Deadline:    "2026-12-31",
		BountyPool:  10000000.00,
	})
	require.NoError(t, err)

	pIntern, err := service.CreateProject(ctx, ownerID, &projects.CreateProjectRequest{
		Title:       "Proyek Internal Khusus Peserta Magang",
		Description: "Pengembangan fitur internal",
		Visibility:  projects.VisibilityInternOnly,
		Capacity:    3,
		Deadline:    "2026-12-31",
		BountyPool:  5000000.00,
	})
	require.NoError(t, err)

	pPrivate, err := service.CreateProject(ctx, ownerID, &projects.CreateProjectRequest{
		Title:       "Proyek Rahasia R&D Inti",
		Description: "Eksperimen arsitektur tertutup",
		Visibility:  projects.VisibilityPrivate,
		Capacity:    2,
		Deadline:    "2026-12-31",
		BountyPool:  15000000.00,
	})
	require.NoError(t, err)

	// Daftarkan memberID sebagai anggota tim proyek PRIVATE
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO project_teams (id, project_id, user_id, project_role, planned_contribution_pct, is_locked)
		VALUES (gen_random_uuid(), $1, $2, 'MEMBER', 100.00, FALSE)
	`, pPrivate.ID, memberID)
	require.NoError(t, err)

	// 2. Matriks visibilitas bursa berbasis izin dinamis (F-RBAC-02, BR-001, BR-003)
	// ALUMNI: hanya PUBLIC
	alumniList, _, err := service.ListProjects(ctx, alumniID, "", "", "", 1, 100)
	require.NoError(t, err)
	require.NotEmpty(t, alumniList)
	for _, p := range alumniList {
		assert.Equal(t, projects.VisibilityPublic, p.Visibility)
	}

	// INTERN: INTERN_ONLY + PUBLIC, tanpa PRIVATE
	internList, _, err := service.ListProjects(ctx, internID, "", "", "", 1, 100)
	require.NoError(t, err)
	require.NotEmpty(t, internList)
	for _, p := range internList {
		assert.NotEqual(t, projects.VisibilityPrivate, p.Visibility)
	}

	// SCANNER: direktori PUBLIC saja (read-only, tanpa PRIVATE)
	scannerList, _, err := service.ListProjects(ctx, scannerID, "", "", "", 1, 100)
	require.NoError(t, err)
	require.NotEmpty(t, scannerList)
	for _, p := range scannerList {
		assert.Equal(t, projects.VisibilityPublic, p.Visibility)
	}

	// Outsider tanpa izin: bursa kosong
	outsiderList, _, err := service.ListProjects(ctx, outsiderID, "", "", "", 1, 100)
	require.NoError(t, err)
	assert.Empty(t, outsiderList)

	// 3. Matriks akses detail proyek PRIVATE
	// Scanner → PRIVATE: DENY
	_, err = service.GetProjectByID(ctx, pPrivate.ID, scannerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "akses ditolak")

	// Outsider → PRIVATE: DENY
	_, err = service.GetProjectByID(ctx, pPrivate.ID, outsiderID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "akses ditolak")

	// Alumni → PRIVATE: DENY (BR-003)
	_, err = service.GetProjectByID(ctx, pPrivate.ID, alumniID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "akses ditolak")

	// Member tim → PRIVATE: PASS
	memberView, err := service.GetProjectByID(ctx, pPrivate.ID, memberID)
	require.NoError(t, err)
	assert.Equal(t, pPrivate.ID, memberView.ID)

	// Owner → PRIVATE: PASS
	pFetched, err := service.GetProjectByID(ctx, pPrivate.ID, ownerID)
	require.NoError(t, err)
	assert.Equal(t, pPrivate.ID, pFetched.ID)

	// Scanner → PUBLIC: PASS
	scannerPublicView, err := service.GetProjectByID(ctx, pPublic.ID, scannerID)
	require.NoError(t, err)
	assert.Equal(t, pPublic.ID, scannerPublicView.ID)

	_ = pIntern
}

// Menguji pengajuan lamaran proyek dan penguncian kuota atomik aman konkurensi (FR-017, T-046).
func TestProjectRegistrationAndAtomicQuotaLock(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := projects.NewRepository(db)
	service := projects.NewService(repo, db, nil, nil, nil)
	ctx := context.Background()

	// Buat Project Owner dan Proyek dengan Kuota 2
	ownerID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'PM Quota Tester', 'ACTIVE')
	`, ownerID, fmt.Sprintf("pm_q_%s@dcisp.internal", ownerID.String()[:8]))
	require.NoError(t, err)

	project, err := service.CreateProject(ctx, ownerID, &projects.CreateProjectRequest{
		Title:       "Proyek Kuota Ketat 2 Peserta",
		Description: "Uji konkurensi kuota tim proyek",
		Visibility:  projects.VisibilityPublic,
		Capacity:    2,
		Deadline:    "2026-12-31",
		BountyPool:  4000000.00,
	})
	require.NoError(t, err)

	// Buat 4 pelamar dengan peran INTERN (izin view marketplace dari data permissions)
	var applicantIDs []uuid.UUID
	var appIDs []uuid.UUID

	for i := 0; i < 4; i++ {
		uID := uuid.New()
		applicantIDs = append(applicantIDs, uID)
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, status)
			VALUES ($1, $2, 'hash_pass', 'Applicant Tester', 'ACTIVE')
		`, uID, fmt.Sprintf("app_%d_%s@dcisp.internal", i, uID.String()[:8]))
		require.NoError(t, err)
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id, scope_id)
			VALUES ($1, '10000000-0000-0000-0000-000000000009', '20000000-0000-0000-0000-000000000007')
		`, uID)
		require.NoError(t, err)

		app, err := service.ApplyProject(ctx, project.ID, uID, &projects.ApplyProjectRequest{
			CoverLetter: "Saya berminat berkontribusi pada proyek ini",
		})
		require.NoError(t, err)
		appIDs = append(appIDs, app.ID)
	}

	t.Cleanup(func() {
		for _, aID := range appIDs {
			_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_applications WHERE id = $1", aID)
		}
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_teams WHERE project_id = $1", project.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM projects WHERE id = $1", project.ID)
		for _, uID := range applicantIDs {
			_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_roles WHERE user_id = $1", uID)
			_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", uID)
		}
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", ownerID)
	})

	// Uji pencegahan duplikasi lamaran
	_, err = service.ApplyProject(ctx, project.ID, applicantIDs[0], &projects.ApplyProjectRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sudah mengajukan")

	// Eksekusi persetujuan pelamar secara konkuren dengan goroutine
	var wg sync.WaitGroup
	var successCount int
	var errorCount int
	var mu sync.Mutex

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := service.ReviewApplication(ctx, appIDs[idx], "ACCEPT")
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	// Kapasitas adalah 2: tepat 2 request harus sukses dan 2 request harus ditolak kuota penuh
	assert.Equal(t, 2, successCount)
	assert.Equal(t, 2, errorCount)

	// Verifikasi nilai accepted_count di database tidak melebihi kapasitas
	updatedProject, err := repo.GetProjectByID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, updatedProject.AcceptedCount)
	assert.True(t, updatedProject.AcceptedCount <= updatedProject.Capacity)
}

// Menguji kalkulasi skor kecocokan keahlian kandidat pelamar yang bersifat informatif tanpa penolakan otomatis (FR-018, BR-015, T-053).
func TestSkillMatchingEngineAdvisoryOnly(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	peopleRepo := people.NewRepository(db)
	projectsRepo := projects.NewRepository(db)
	service := projects.NewService(projectsRepo, db, nil, nil, nil)
	ctx := context.Background()

	// 1. Buat master skill
	s1 := &people.Skill{ID: uuid.New(), Name: fmt.Sprintf("Skill Go %d", time.Now().UnixNano()%100000), Category: "BACKEND"}
	s2 := &people.Skill{ID: uuid.New(), Name: fmt.Sprintf("Skill React %d", time.Now().UnixNano()%100000), Category: "FRONTEND"}
	s3 := &people.Skill{ID: uuid.New(), Name: fmt.Sprintf("Skill Docker %d", time.Now().UnixNano()%100000), Category: "DEVOPS"}
	require.NoError(t, peopleRepo.CreateSkill(ctx, s1))
	require.NoError(t, peopleRepo.CreateSkill(ctx, s2))
	require.NoError(t, peopleRepo.CreateSkill(ctx, s3))

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM skills WHERE id IN ($1, $2, $3)", s1.ID, s2.ID, s3.ID)
	})

	// 2. Setup user kandidat dan tetapkan skill s1 dan s2
	candidateID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Candidate Match Tester', 'ACTIVE')
	`, candidateID, fmt.Sprintf("cand_%s@dcisp.internal", candidateID.String()[:8]))
	require.NoError(t, err)

	require.NoError(t, peopleRepo.UpsertUserSkill(ctx, &people.UserSkill{ID: uuid.New(), UserID: candidateID, SkillID: s1.ID, ProficiencyLevel: 4}))
	require.NoError(t, peopleRepo.UpsertUserSkill(ctx, &people.UserSkill{ID: uuid.New(), UserID: candidateID, SkillID: s2.ID, ProficiencyLevel: 3}))

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_skills WHERE user_id = $1", candidateID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", candidateID)
	})

	// 3. Uji kecocokan: Proyek membutuhkan s1 dan s2 -> Harus 100.00%
	scoreFull := service.CalculateSkillMatchPercentage(ctx, []uuid.UUID{s1.ID, s2.ID}, candidateID)
	assert.Equal(t, 100.00, scoreFull)

	// 4. Uji kecocokan: Proyek membutuhkan s1, s2, dan s3 (s3 tidak dimiliki kandidat) -> 2 dari 3 = 66.67%
	scorePartial := service.CalculateSkillMatchPercentage(ctx, []uuid.UUID{s1.ID, s2.ID, s3.ID}, candidateID)
	assert.Equal(t, 66.67, scorePartial)

	// 5. Uji kecocokan: Proyek tanpa syarat skill -> Harus 100.00%
	scoreEmpty := service.CalculateSkillMatchPercentage(ctx, []uuid.UUID{}, candidateID)
	assert.Equal(t, 100.00, scoreEmpty)
}
