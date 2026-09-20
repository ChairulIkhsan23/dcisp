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

// Menguji seluruh siklus hidup tugas kanban, penyerahan laporan kerja dengan bukti deliverable, dan pengesahan reviewer (FR-021, FR-022, FR-023, BRULE-PRJ-003).
func TestTaskKanbanAndWorkReportSubmission(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := projects.NewRepository(db)
	service := projects.NewService(repo, db, nil, nil, nil)
	ctx := context.Background()

	// 1. Setup Project Owner, Member, dan Reviewer
	ownerID := uuid.New()
	memberID := uuid.New()
	reviewerID := uuid.New()

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'PM Owner', 'ACTIVE'),
		       ($3, $4, 'hash_pass', 'Member Worker', 'ACTIVE'),
		       ($5, $6, 'hash_pass', 'Reviewer Oracle', 'ACTIVE')
	`, ownerID, fmt.Sprintf("owner_%s@dcisp.internal", ownerID.String()[:8]),
		memberID, fmt.Sprintf("member_%s@dcisp.internal", memberID.String()[:8]),
		reviewerID, fmt.Sprintf("rev_%s@dcisp.internal", reviewerID.String()[:8]))
	require.NoError(t, err)

	project, err := service.CreateProject(ctx, ownerID, &projects.CreateProjectRequest{
		Title:       "Proyek Kanban & Deliverable",
		Description: "Uji coba alur kanban dan laporan deliverable",
		Visibility:  projects.VisibilityPublic,
		Capacity:    3,
		Deadline:    "2026-12-31",
		BountyPool:  6000000.00,
	})
	require.NoError(t, err)

	// Tambahkan member ke project_teams
	memberTeam := &projects.ProjectTeamMember{
		ID:                     uuid.New(),
		ProjectID:              project.ID,
		UserID:                 memberID,
		ProjectRole:            projects.ProjectRoleMember,
		PlannedContributionPct: 50.00,
	}
	require.NoError(t, repo.AddTeamMember(ctx, memberTeam))

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM evidence WHERE submission_id IN (SELECT id FROM work_reports WHERE user_id = $1)", memberID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM work_reports WHERE user_id = $1", memberID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM tasks WHERE project_id = $1", project.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM milestones WHERE project_id = $1", project.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_teams WHERE project_id = $1", project.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM projects WHERE id = $1", project.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id IN ($1, $2, $3)", ownerID, memberID, reviewerID)
	})

	// 2. Buat Milestone dengan bobot valid 40.00% (FR-020, T-048)
	milestone, err := service.CreateMilestone(ctx, project.ID, &projects.CreateMilestoneRequest{
		Title:     "Milestone 1 — Arsitektur Backend",
		Deadline:  "2026-10-31",
		WeightPct: 40.00,
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, milestone.ID)

	// Uji penolakan bobot melebihi 100%: buat milestone kedua dengan bobot 70% (40 + 70 = 110%)
	_, err = service.CreateMilestone(ctx, project.ID, &projects.CreateMilestoneRequest{
		Title:     "Milestone 2 — Bobot Berlebih",
		Deadline:  "2026-11-30",
		WeightPct: 70.00,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "melebihi batas 100%")

	// 3. Buat Tugas Kanban (FR-021, T-049)
	task, err := service.CreateTask(ctx, project.ID, &projects.CreateTaskRequest{
		MilestoneID:      &milestone.ID,
		AssigneeID:       &memberID,
		Title:            "Implementasi Endpoint Autentikasi JWT",
		Description:      func(s string) *string { return &s }("Buat endpoint login dan refresh token"),
		EstimatedHours:   8.0,
		DifficultyWeight: 3,
		Priority:         projects.TaskPriorityHigh,
		Deadline:         "2026-10-15",
	})
	require.NoError(t, err)
	assert.Equal(t, projects.TaskStatusTodo, task.Status)

	// 4. Transisi Mesin Status: TODO -> IN_PROGRESS
	task, err = service.ChangeTaskStatus(ctx, task.ID, projects.TaskStatusInProgress)
	require.NoError(t, err)
	assert.Equal(t, projects.TaskStatusInProgress, task.Status)

	// Uji penolakan langsung ke COMPLETED tanpa persetujuan laporan kerja (BRULE-PRJ-003)
	_, err = service.ChangeTaskStatus(ctx, task.ID, projects.TaskStatusCompleted)
	assert.Error(t, err)

	// 5. Penyerahan Laporan Pekerjaan (Work Report & Evidence - FR-022, FR-023, T-050, T-051)
	reportReq := &projects.SubmitWorkReportRequest{
		ProgressPercentage: 100,
		WhatIDid:           "Mengimplementasikan JWT auth middleware dan unit test otentikasi lengkap",
		EvidenceType:       projects.EvidenceTypeGitCommit,
		EvidenceURLOrKey:   "https://github.com/company/repo/commit/abc1234567890",
	}
	report, err := service.SubmitWorkReport(ctx, task.ID, memberID, reportReq)
	require.NoError(t, err)
	assert.Equal(t, projects.ReportStatusSubmitted, report.Status)

	// Verifikasi status tugas otomatis beralih ke IN_REVIEW
	taskInReview, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, projects.TaskStatusInReview, taskInReview.Status)

	// 6. Penelaahan Reviewer: Minta Revisi (REVISION_REQUIRED)
	reviewedReport, err := service.ReviewWorkReport(ctx, report.ID, "REVISION_REQUIRED")
	require.NoError(t, err)
	assert.Equal(t, projects.ReportStatusRevisionRequired, reviewedReport.Status)

	// Verifikasi status tugas otomatis kembali ke IN_PROGRESS
	taskRevising, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, projects.TaskStatusInProgress, taskRevising.Status)

	// 7. Penyerahan ulang laporan pekerjaan setelah revisi
	report2, err := service.SubmitWorkReport(ctx, task.ID, memberID, reportReq)
	require.NoError(t, err)

	// 8. Penelaahan Reviewer: Setujui Laporan (APPROVE)
	approvedReport, err := service.ReviewWorkReport(ctx, report2.ID, "APPROVE")
	require.NoError(t, err)
	assert.Equal(t, projects.ReportStatusApproved, approvedReport.Status)

	// Verifikasi status tugas sah menjadi COMPLETED (BRULE-PRJ-003)
	taskCompleted, err := repo.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, projects.TaskStatusCompleted, taskCompleted.Status)
}

// Menguji rekonsiliasi model kontribusi tiga lapis: Planned, komputasi Actual, dan pengesahan Final % (FR-019, FR-024, BR-016, T-047, T-052).
func TestThreeLayerContributionEngine(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := projects.NewRepository(db)
	service := projects.NewService(repo, db, nil, nil, nil)
	ctx := context.Background()

	// 1. Setup Project, PM, Supervisor, dan 2 Anggota Tim (A & B)
	ownerID := uuid.New()
	supervisorID := uuid.New()
	memberA := uuid.New()
	memberB := uuid.New()

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'PM Owner', 'ACTIVE'),
		       ($3, $4, 'hash_pass', 'Supervisor Guild', 'ACTIVE'),
		       ($5, $6, 'hash_pass', 'Member A', 'ACTIVE'),
		       ($7, $8, 'hash_pass', 'Member B', 'ACTIVE')
	`, ownerID, fmt.Sprintf("owner_%s@dcisp.internal", ownerID.String()[:8]),
		supervisorID, fmt.Sprintf("spv_%s@dcisp.internal", supervisorID.String()[:8]),
		memberA, fmt.Sprintf("ma_%s@dcisp.internal", memberA.String()[:8]),
		memberB, fmt.Sprintf("mb_%s@dcisp.internal", memberB.String()[:8]))
	require.NoError(t, err)

	project, err := service.CreateProject(ctx, ownerID, &projects.CreateProjectRequest{
		Title:       "Proyek Rekonsiliasi Tiga Lapis",
		Description: "Uji coba Three-Layer Contribution",
		Visibility:  projects.VisibilityPublic,
		Capacity:    3,
		Deadline:    "2026-12-31",
		BountyPool:  10000000.00,
	})
	require.NoError(t, err)

	// Daftarkan Member A dan Member B ke tim proyek
	require.NoError(t, repo.AddTeamMember(ctx, &projects.ProjectTeamMember{
		ID: uuid.New(), ProjectID: project.ID, UserID: memberA, ProjectRole: projects.ProjectRoleMember,
	}))
	require.NoError(t, repo.AddTeamMember(ctx, &projects.ProjectTeamMember{
		ID: uuid.New(), ProjectID: project.ID, UserID: memberB, ProjectRole: projects.ProjectRoleMember,
	}))

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM tasks WHERE project_id = $1", project.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_teams WHERE project_id = $1", project.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM projects WHERE id = $1", project.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id IN ($1, $2, $3, $4)", ownerID, supervisorID, memberA, memberB)
	})

	// 2. LAYER 1: Penetapan Planned Contribution % (T-047)
	// Uji penolakan jika total tidak 100% (misal 60 + 30 = 90%)
	err = service.FinalizePlannedContribution(ctx, project.ID, &projects.FinalizePlannedContributionRequest{
		Contributions: []projects.MemberPlannedContribution{
			{UserID: memberA, PlannedContributionPct: 60.00},
			{UserID: memberB, PlannedContributionPct: 30.00},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tepat 100.00%")

	// Penetapan valid: 60% dan 40% (total = 100.00%)
	err = service.FinalizePlannedContribution(ctx, project.ID, &projects.FinalizePlannedContributionRequest{
		Contributions: []projects.MemberPlannedContribution{
			{UserID: memberA, PlannedContributionPct: 60.00},
			{UserID: memberB, PlannedContributionPct: 40.00},
		},
	})
	require.NoError(t, err)

	// 3. LAYER 2: Komputasi Actual Contribution % berbasis data tugas COMPLETED
	// Buat tugas selesai:
	// Member A menyelesaikan tugas dengan bobot kesulitan 4 dan durasi 10 jam (Skor = 40)
	// Member B menyelesaikan tugas dengan bobot kesulitan 2 dan durasi 5 jam (Skor = 10)
	// Total Skor = 50 -> Actual % Member A = 40/50 = 80.00%, Member B = 10/50 = 20.00%
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO tasks (id, project_id, assignee_id, title, estimated_hours, difficulty_weight, priority, deadline, status)
		VALUES (gen_random_uuid(), $1, $2, 'Task A Selesai', 10.0, 4, 'HIGH', CURRENT_TIMESTAMP, 'COMPLETED'),
		       (gen_random_uuid(), $1, $3, 'Task B Selesai', 5.0, 2, 'MEDIUM', CURRENT_TIMESTAMP, 'COMPLETED')
	`, project.ID, memberA, memberB)
	require.NoError(t, err)

	actualResults, err := service.CalculateActualContribution(ctx, project.ID)
	require.NoError(t, err)

	var actualA, actualB float64
	for _, m := range actualResults {
		if m.UserID == memberA {
			actualA = *m.ActualContributionPct
		}
		if m.UserID == memberB {
			actualB = *m.ActualContributionPct
		}
	}
	assert.Equal(t, 80.00, actualA)
	assert.Equal(t, 20.00, actualB)

	// 4. LAYER 3: Pengesahan Final Contribution % oleh Supervisor (BR-016)
	// Uji penolakan jika total tidak 100% (misal 70 + 20 = 90%)
	err = service.FinalizeContribution(ctx, project.ID, &projects.FinalizeContributionRequest{
		FinalContributions: []projects.MemberFinalContribution{
			{UserID: memberA, FinalContributionPct: 70.00},
			{UserID: memberB, FinalContributionPct: 20.00},
		},
		EvaluationNotes: "Catatan evaluasi tidak seimbang",
	}, supervisorID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tepat 100.00%")

	// Pengesahan sah: Supervisor memutuskan 75.00% dan 25.00% (total = 100.00%)
	err = service.FinalizeContribution(ctx, project.ID, &projects.FinalizeContributionRequest{
		FinalContributions: []projects.MemberFinalContribution{
			{UserID: memberA, FinalContributionPct: 75.00},
			{UserID: memberB, FinalContributionPct: 25.00},
		},
		EvaluationNotes: "Member A memberikan kontribusi arsitektur dominan terkonfirmasi",
	}, supervisorID)
	require.NoError(t, err)

	// Verifikasi bahwa status tim terkunci (is_locked = true)
	memberAFinal, err := repo.GetTeamMember(ctx, project.ID, memberA)
	require.NoError(t, err)
	assert.True(t, memberAFinal.IsLocked)
	assert.Equal(t, 75.00, *memberAFinal.FinalContributionPct)
	assert.Equal(t, 60.00, memberAFinal.PlannedContributionPct)
	assert.Equal(t, 80.00, *memberAFinal.ActualContributionPct)
}
