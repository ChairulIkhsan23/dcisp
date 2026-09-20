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

	// Setup Project Owner
	ownerID := uuid.New()
	ownerEmail := fmt.Sprintf("pm_%s@dcisp.internal", ownerID.String()[:8])
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Project Manager Dimas', 'ACTIVE')
	`, ownerID, ownerEmail)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_teams WHERE user_id = $1", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM projects WHERE owner_id = $1", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", ownerID)
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

	// 2. Uji visibilitas untuk peran ALUMNI (Hanya boleh melihat PUBLIC - BR-003)
	alumniList, _, err := service.ListProjects(ctx, "ALUMNI", "", "", "", 1, 20)
	require.NoError(t, err)
	for _, p := range alumniList {
		assert.Equal(t, projects.VisibilityPublic, p.Visibility)
	}

	// 3. Uji visibilitas untuk peran INTERN (Boleh melihat INTERN_ONLY dan PUBLIC)
	internList, _, err := service.ListProjects(ctx, "INTERN", "", "", "", 1, 20)
	require.NoError(t, err)
	for _, p := range internList {
		assert.NotEqual(t, projects.VisibilityPrivate, p.Visibility)
	}

	// 4. Uji isolasi akses proyek PRIVATE: pengguna luar ditolak akses detailnya
	outsiderID := uuid.New()
	_, err = service.GetProjectByID(ctx, pPrivate.ID, outsiderID, "INTERN")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "akses ditolak")

	// Pemilik proyek diizinkan mengakses proyek PRIVATE
	pFetched, err := service.GetProjectByID(ctx, pPrivate.ID, ownerID, "PROJECT_MANAGER")
	require.NoError(t, err)
	assert.Equal(t, pPrivate.ID, pFetched.ID)

	_ = pPublic
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

	// Buat 4 pelamar
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

		app, err := service.ApplyProject(ctx, project.ID, uID, &projects.ApplyProjectRequest{
			CoverLetter: "Saya berminat berkontribusi pada proyek ini",
		}, "INTERN")
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
			_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", uID)
		}
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", ownerID)
	})

	// Uji pencegahan duplikasi lamaran
	_, err = service.ApplyProject(ctx, project.ID, applicantIDs[0], &projects.ApplyProjectRequest{}, "INTERN")
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
