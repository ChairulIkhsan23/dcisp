package tests

import (
	"context"
	"fmt"
	"testing"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/projects"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji alur lengkap siklus hidup proyek dari pembuatan hingga pengesahan kontribusi akhir (E2E Phase 5).
func TestProjectCompleteLifecycleE2E(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := projects.NewRepository(db)
	service := projects.NewService(repo, db, nil, nil, nil)
	ctx := context.Background()

	// 1. Setup Aktor: Project Owner, Supervisor, dan 2 Pelamar/Anggota
	ownerID := uuid.New()
	supervisorID := uuid.New()
	applicant1ID := uuid.New()
	applicant2ID := uuid.New()

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'PM Dimas', 'ACTIVE'),
		       ($3, $4, 'hash_pass', 'Supervisor Citra', 'ACTIVE'),
		       ($5, $6, 'hash_pass', 'Intern Andi', 'ACTIVE'),
		       ($7, $8, 'hash_pass', 'Intern Budi', 'ACTIVE')
	`, ownerID, fmt.Sprintf("pm_e2e_%s@dcisp.internal", ownerID.String()[:8]),
		supervisorID, fmt.Sprintf("spv_e2e_%s@dcisp.internal", supervisorID.String()[:8]),
		applicant1ID, fmt.Sprintf("app1_e2e_%s@dcisp.internal", applicant1ID.String()[:8]),
		applicant2ID, fmt.Sprintf("app2_e2e_%s@dcisp.internal", applicant2ID.String()[:8]))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM evidence WHERE submission_id IN (SELECT id FROM work_reports WHERE user_id IN ($1, $2))", applicant1ID, applicant2ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM work_reports WHERE user_id IN ($1, $2)", applicant1ID, applicant2ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM tasks WHERE project_id IN (SELECT id FROM projects WHERE owner_id = $1)", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM milestones WHERE project_id IN (SELECT id FROM projects WHERE owner_id = $1)", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_applications WHERE project_id IN (SELECT id FROM projects WHERE owner_id = $1)", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_teams WHERE project_id IN (SELECT id FROM projects WHERE owner_id = $1)", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM projects WHERE owner_id = $1", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id IN ($1, $2, $3, $4)", ownerID, supervisorID, applicant1ID, applicant2ID)
	})

	// 2. Create Project dengan Visibilitas INTERN_ONLY dan Kuota 2
	project, err := service.CreateProject(ctx, ownerID, &projects.CreateProjectRequest{
		Title:       "Proyek E2E DCISP v1.0",
		Description: "Pengujian end-to-end menyeluruh siklus proyek",
		Visibility:  projects.VisibilityInternOnly,
		Capacity:    2,
		Deadline:    "2026-12-31",
		BountyPool:  10000000.00,
	})
	require.NoError(t, err)

	// 3. Pelamar 1 dan Pelamar 2 Mengajukan Lamaran
	app1, err := service.ApplyProject(ctx, project.ID, applicant1ID, &projects.ApplyProjectRequest{CoverLetter: "Siap bekerja backend"}, "INTERN")
	require.NoError(t, err)
	app2, err := service.ApplyProject(ctx, project.ID, applicant2ID, &projects.ApplyProjectRequest{CoverLetter: "Siap bekerja frontend"}, "INTERN")
	require.NoError(t, err)

	// 4. PM Menerima Pelamar 1 dan Pelamar 2 (Penguncian Kuota Atomik)
	_, err = service.ReviewApplication(ctx, app1.ID, "ACCEPT")
	require.NoError(t, err)
	_, err = service.ReviewApplication(ctx, app2.ID, "ACCEPT")
	require.NoError(t, err)

	// Verifikasi kuota terkunci penuh (accepted_count = capacity = 2)
	lockedProj, err := repo.GetProjectByID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, lockedProj.AcceptedCount)

	// 5. Penetapan Planned Contribution % Seluruh Anggota Tim (Total 100.00%)
	err = service.FinalizePlannedContribution(ctx, project.ID, &projects.FinalizePlannedContributionRequest{
		Contributions: []projects.MemberPlannedContribution{
			{UserID: applicant1ID, PlannedContributionPct: 60.00},
			{UserID: applicant2ID, PlannedContributionPct: 40.00},
		},
	})
	require.NoError(t, err)

	// 6. Pembuatan Milestones dengan Total Bobot Tepat 100.00%
	m1, err := service.CreateMilestone(ctx, project.ID, &projects.CreateMilestoneRequest{
		Title:     "Fase 1 — Desain & Fondasi",
		Deadline:  "2026-10-31",
		WeightPct: 50.00,
	})
	require.NoError(t, err)

	m2, err := service.CreateMilestone(ctx, project.ID, &projects.CreateMilestoneRequest{
		Title:     "Fase 2 — Integrasi & Pengujian",
		Deadline:  "2026-11-30",
		WeightPct: 50.00,
	})
	require.NoError(t, err)

	totalWeight, err := repo.GetTotalMilestoneWeight(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, 100.00, totalWeight)

	// 7. Pembuatan dan Penugasan Tugas Kanban
	t1, err := service.CreateTask(ctx, project.ID, &projects.CreateTaskRequest{
		MilestoneID:      &m1.ID,
		AssigneeID:       &applicant1ID,
		Title:            "Membangun API Arsitektur",
		EstimatedHours:   20.0,
		DifficultyWeight: 4, // Skor = 20 * 4 = 80
		Priority:         projects.TaskPriorityHigh,
		Deadline:         "2026-10-15",
	})
	require.NoError(t, err)

	t2, err := service.CreateTask(ctx, project.ID, &projects.CreateTaskRequest{
		MilestoneID:      &m2.ID,
		AssigneeID:       &applicant2ID,
		Title:            "Membangun Tampilan Antarmuka",
		EstimatedHours:   10.0,
		DifficultyWeight: 2, // Skor = 10 * 2 = 20
		Priority:         projects.TaskPriorityMedium,
		Deadline:         "2026-11-15",
	})
	require.NoError(t, err)

	// 8. Eksekusi Siklus Tugas 1 (Pelamar 1)
	_, err = service.ChangeTaskStatus(ctx, t1.ID, projects.TaskStatusInProgress)
	require.NoError(t, err)
	rep1, err := service.SubmitWorkReport(ctx, t1.ID, applicant1ID, &projects.SubmitWorkReportRequest{
		ProgressPercentage: 100,
		WhatIDid:           "Menyelesaikan implementasi modul backend secara lengkap",
		EvidenceType:       projects.EvidenceTypeGitCommit,
		EvidenceURLOrKey:   "https://github.com/company/repo/commit/e2e123456789",
	})
	require.NoError(t, err)
	_, err = service.ReviewWorkReport(ctx, rep1.ID, "APPROVE")
	require.NoError(t, err)

	// 9. Eksekusi Siklus Tugas 2 (Pelamar 2)
	_, err = service.ChangeTaskStatus(ctx, t2.ID, projects.TaskStatusInProgress)
	require.NoError(t, err)
	rep2, err := service.SubmitWorkReport(ctx, t2.ID, applicant2ID, &projects.SubmitWorkReportRequest{
		ProgressPercentage: 100,
		WhatIDid:           "Menyelesaikan antarmuka halaman web sesuai panduan desain",
		EvidenceType:       projects.EvidenceTypeURL,
		EvidenceURLOrKey:   "https://staging.dcisp.internal/preview/frontend",
	})
	require.NoError(t, err)
	_, err = service.ReviewWorkReport(ctx, rep2.ID, "APPROVE")
	require.NoError(t, err)

	// 10. Komputasi Layer 2 (Actual Contribution %)
	// Skor Pelamar 1 = 80, Skor Pelamar 2 = 20. Total = 100 -> Actual % = 80.00% dan 20.00%
	actualMembers, err := service.CalculateActualContribution(ctx, project.ID)
	require.NoError(t, err)

	var actual1, actual2 float64
	for _, m := range actualMembers {
		if m.UserID == applicant1ID {
			actual1 = *m.ActualContributionPct
		}
		if m.UserID == applicant2ID {
			actual2 = *m.ActualContributionPct
		}
	}
	assert.Equal(t, 80.00, actual1)
	assert.Equal(t, 20.00, actual2)

	// 11. Pengesahan Layer 3 (Final Contribution %) oleh Supervisor (Total 100.00%)
	err = service.FinalizeContribution(ctx, project.ID, &projects.FinalizeContributionRequest{
		FinalContributions: []projects.MemberFinalContribution{
			{UserID: applicant1ID, FinalContributionPct: 75.00},
			{UserID: applicant2ID, FinalContributionPct: 25.00},
		},
		EvaluationNotes: "Disahkan berdasarkan pertimbangan usaha ekstra pada penyelesaian arsitektur",
	}, supervisorID)
	require.NoError(t, err)

	// Verifikasi penguncian akhir kontribusi tim
	finalMember1, err := repo.GetTeamMember(ctx, project.ID, applicant1ID)
	require.NoError(t, err)
	assert.True(t, finalMember1.IsLocked)
	assert.Equal(t, 75.00, *finalMember1.FinalContributionPct)
	assert.Equal(t, 60.00, finalMember1.PlannedContributionPct)
	assert.Equal(t, 80.00, *finalMember1.ActualContributionPct)

	finalMember2, err := repo.GetTeamMember(ctx, project.ID, applicant2ID)
	require.NoError(t, err)
	assert.True(t, finalMember2.IsLocked)
	assert.Equal(t, 25.00, *finalMember2.FinalContributionPct)
	assert.Equal(t, 40.00, finalMember2.PlannedContributionPct)
	assert.Equal(t, 20.00, *finalMember2.ActualContributionPct)

	// Verifikasi status proyek menjadi COMPLETED
	completedProject, err := repo.GetProjectByID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, projects.ProjectStatusCompleted, completedProject.Status)
}
