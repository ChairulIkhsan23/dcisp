package projects

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/documents"
	"dcisp/backend/internal/modules/system"
	"dcisp/backend/internal/shared/eventbus"
	"github.com/google/uuid"
)

type Service struct {
	repo           *Repository
	db             *database.PostgresDB
	storageService *documents.StorageService
	auditService   *system.AuditService
	eventBus       *eventbus.EventBus
}

// Menginisialisasi instance baru service projects and tasks management.
func NewService(repo *Repository, db *database.PostgresDB, storage *documents.StorageService, audit *system.AuditService, bus *eventbus.EventBus) *Service {
	return &Service{
		repo:           repo,
		db:             db,
		storageService: storage,
		auditService:   audit,
		eventBus:       bus,
	}
}

// ============================================================================
// 1. PROJECT MARKETPLACE (FR-016, T-045)
// ============================================================================

// Membuat proyek baru pada bursa kerja dan menetapkan pemilik proyek pada tim.
func (s *Service) CreateProject(ctx context.Context, ownerID uuid.UUID, req *CreateProjectRequest) (*Project, error) {
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		deadline, err = time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			return nil, fmt.Errorf("format tenggat waktu tidak valid (harus ISO8601/RFC3339): %w", err)
		}
	}
	if deadline.Before(time.Now().UTC()) {
		return nil, errors.New("tenggat waktu proyek tidak boleh berada di masa lalu")
	}

	vis := strings.ToUpper(req.Visibility)
	if vis != VisibilityInternOnly && vis != VisibilityPublic && vis != VisibilityPrivate {
		return nil, fmt.Errorf("visibilitas proyek tidak valid: %s (diharapkan INTERN_ONLY, PUBLIC, atau PRIVATE)", req.Visibility)
	}

	status := ProjectStatusPublished
	if req.Status != "" {
		status = strings.ToUpper(req.Status)
	}

	skills := req.RequiredSkills
	if skills == nil {
		skills = []uuid.UUID{}
	}

	project := &Project{
		ID:             uuid.New(),
		Title:          req.Title,
		Description:    req.Description,
		OwnerID:        ownerID,
		Visibility:     vis,
		RequiredSkills: skills,
		Capacity:       req.Capacity,
		AcceptedCount:  0,
		Deadline:       deadline,
		BountyPool:     req.BountyPool,
		Status:         status,
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi pembuatan proyek: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.CreateProject(ctx, project); err != nil {
		return nil, err
	}

	// Daftarkan pembuat proyek sebagai OWNER pada tim proyek
	ownerMember := &ProjectTeamMember{
		ID:                     uuid.New(),
		ProjectID:              project.ID,
		UserID:                 ownerID,
		ProjectRole:            ProjectRoleOwner,
		Responsibility:         func(s string) *string { return &s }("Project Initiator & Owner"),
		PlannedContributionPct: 0.00,
		IsLocked:               false,
	}
	if err := s.repo.AddTeamMemberTx(ctx, tx, ownerMember); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit pembuatan proyek: %w", err)
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "CREATE_PROJECT",
			ResourceType: "PROJECT",
			ResourceID:   project.ID.String(),
			NewState: map[string]interface{}{
				"title":       project.Title,
				"visibility":  project.Visibility,
				"bounty_pool": project.BountyPool,
				"capacity":    project.Capacity,
			},
		})
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "project.created",
			AggregateID: project.ID.String(),
			Payload: map[string]interface{}{
				"project_id":  project.ID.String(),
				"title":       project.Title,
				"visibility":  project.Visibility,
				"bounty_pool": project.BountyPool,
			},
		})
	}

	return project, nil
}

// Mengambil daftar proyek bursa dengan penegakan batasan visibilitas berdasarkan peran pengguna (FR-016, BR-003).
func (s *Service) ListProjects(ctx context.Context, requesterRole string, status, search, requestedVis string, page, limit int) ([]Project, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Penegakan aturan visibilitas (BR-003): Alumni hanya boleh mengakses proyek bertipe PUBLIC
	var allowedVisibilities []string
	if strings.EqualFold(requesterRole, "ALUMNI") {
		allowedVisibilities = []string{VisibilityPublic}
	} else if strings.EqualFold(requesterRole, "INTERN") {
		allowedVisibilities = []string{VisibilityInternOnly, VisibilityPublic}
	} else {
		// Admin, PM, Super Admin dapat melihat semua
		if requestedVis != "" {
			allowedVisibilities = []string{strings.ToUpper(requestedVis)}
		}
	}

	return s.repo.ListProjects(ctx, allowedVisibilities, status, search, limit, offset)
}

// Mengambil detail satu proyek setelah memvalidasi izin visibilitas pemohon.
func (s *Service) GetProjectByID(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole string) (*Project, error) {
	p, err := s.repo.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("proyek tidak ditemukan")
	}

	// Pemeriksaan hak akses visibilitas
	if strings.EqualFold(requesterRole, "ALUMNI") && p.Visibility != VisibilityPublic {
		return nil, errors.New("akses ditolak: alumni hanya dapat mengakses proyek dengan visibilitas PUBLIC (BR-003)")
	}
	if p.Visibility == VisibilityPrivate {
		isMember, _ := s.repo.GetTeamMember(ctx, p.ID, requesterID)
		if isMember == nil && p.OwnerID != requesterID && !strings.Contains(strings.ToUpper(requesterRole), "ADMIN") {
			return nil, errors.New("akses ditolak: proyek ini bersifat PRIVATE")
		}
	}

	return p, nil
}

// Memperbarui informasi proyek master.
func (s *Service) UpdateProject(ctx context.Context, id uuid.UUID, req *UpdateProjectRequest, requesterID uuid.UUID) (*Project, error) {
	p, err := s.repo.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("proyek tidak ditemukan")
	}

	if req.Title != "" {
		p.Title = req.Title
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Visibility != "" {
		p.Visibility = strings.ToUpper(req.Visibility)
	}
	if req.RequiredSkills != nil {
		p.RequiredSkills = req.RequiredSkills
	}
	if req.Capacity != nil {
		if *req.Capacity < p.AcceptedCount {
			return nil, fmt.Errorf("kapasitas baru (%d) tidak boleh lebih kecil dari peserta yang sudah diterima (%d)", *req.Capacity, p.AcceptedCount)
		}
		p.Capacity = *req.Capacity
	}
	if req.Deadline != "" {
		parsedDate, err := time.Parse(time.RFC3339, req.Deadline)
		if err != nil {
			return nil, fmt.Errorf("format tenggat waktu tidak valid: %w", err)
		}
		p.Deadline = parsedDate
	}
	if req.BountyPool != nil {
		p.BountyPool = *req.BountyPool
	}
	if req.Status != "" {
		p.Status = strings.ToUpper(req.Status)
	}

	if err := s.repo.UpdateProject(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// ============================================================================
// 2. PROJECT APPLICATION & QUOTA (FR-017, T-046)
// ============================================================================

// Mengajukan lamaran pada proyek dengan kalkulasi skor kecocokan keahlian informatif (advisory only).
func (s *Service) ApplyProject(ctx context.Context, projectID, userID uuid.UUID, req *ApplyProjectRequest, userRole string) (*ProjectApplication, error) {
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("proyek tidak ditemukan")
	}

	if project.Status != ProjectStatusPublished {
		return nil, fmt.Errorf("proyek saat ini tidak membuka lamaran (status: %s)", project.Status)
	}

	// Pemeriksaan batasan visibilitas peran alumni
	if strings.EqualFold(userRole, "ALUMNI") && project.Visibility != VisibilityPublic {
		return nil, errors.New("alumni hanya dapat melamar pada proyek dengan visibilitas PUBLIC (BR-003)")
	}

	// Cek apakah sudah pernah melamar
	existingApp, _ := s.repo.GetApplication(ctx, projectID, userID)
	if existingApp != nil {
		return nil, errors.New("anda sudah mengajukan lamaran pada proyek ini sebelumnya")
	}

	// Cek kapasitas kuota awal
	if project.AcceptedCount >= project.Capacity {
		return nil, errors.New("kuota peserta proyek telah penuh")
	}

	// Hitung skor kecocokan keahlian secara informatif (T-053 - FR-018)
	matchScore := s.CalculateSkillMatchPercentage(ctx, project.RequiredSkills, userID)

	coverLetter := req.CoverLetter
	app := &ProjectApplication{
		ID:                   uuid.New(),
		ProjectID:            projectID,
		UserID:               userID,
		CoverLetter:          &coverLetter,
		SkillMatchPercentage: &matchScore,
		Status:               AppStatusApplied,
	}

	if err := s.repo.CreateApplication(ctx, app); err != nil {
		return nil, err
	}

	return app, nil
}

// Meninjau dan menyetujui atau menolak lamaran proyek dengan penguncian kuota atomik aman konkurensi (FR-017).
func (s *Service) ReviewApplication(ctx context.Context, appID uuid.UUID, action string) (*ProjectApplication, error) {
	app, err := s.repo.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("lamaran tidak ditemukan")
	}

	upperAction := strings.ToUpper(action)
	if upperAction != "ACCEPT" && upperAction != "REJECT" {
		return nil, errors.New("aksi tidak valid, diharapkan 'ACCEPT' atau 'REJECT'")
	}

	if upperAction == "REJECT" {
		tx, err := s.db.Pool.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)

		if err := s.repo.UpdateApplicationStatusTx(ctx, tx, app.ID, AppStatusRejected); err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		app.Status = AppStatusRejected
		return app, nil
	}

	// Transaksi atomik SERIALIZABLE / FOR UPDATE untuk penguncian kuota (T-046)
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi penerimaan pelamar: %w", err)
	}
	defer tx.Rollback(ctx)

	// Kunci baris proyek dengan FOR UPDATE
	lockedProject, err := s.repo.LockProjectForUpdateTx(ctx, tx, app.ProjectID)
	if err != nil {
		return nil, err
	}
	if lockedProject == nil {
		return nil, errors.New("proyek tidak ditemukan")
	}

	if lockedProject.AcceptedCount >= lockedProject.Capacity {
		return nil, fmt.Errorf("kuota anggota tim proyek '%s' telah penuh (%d/%d)", lockedProject.Title, lockedProject.AcceptedCount, lockedProject.Capacity)
	}

	// Tingkatkan jumlah peserta diterima
	if err := s.repo.IncrementAcceptedCountTx(ctx, tx, app.ProjectID); err != nil {
		return nil, err
	}

	// Perbarui status lamaran ke ACCEPTED
	if err := s.repo.UpdateApplicationStatusTx(ctx, tx, app.ID, AppStatusAccepted); err != nil {
		return nil, err
	}

	// Tambahkan peserta sebagai anggota tim proyek (MEMBER)
	teamMember := &ProjectTeamMember{
		ID:                     uuid.New(),
		ProjectID:              app.ProjectID,
		UserID:                 app.UserID,
		ProjectRole:            ProjectRoleMember,
		Responsibility:         func(s string) *string { return &s }("Project Team Member"),
		PlannedContributionPct: 0.00,
		IsLocked:               false,
	}
	if err := s.repo.AddTeamMemberTx(ctx, tx, teamMember); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit penerimaan pelamar: %w", err)
	}

	app.Status = AppStatusAccepted

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "project.member_joined",
			AggregateID: app.ProjectID.String(),
			Payload: map[string]interface{}{
				"project_id": app.ProjectID.String(),
				"user_id":    app.UserID.String(),
			},
		})
	}

	return app, nil
}

// ============================================================================
// 3. PROJECT TEAMS & THREE-LAYER CONTRIBUTION (FR-019, FR-024, T-047, T-052)
// ============================================================================

// Memfinalisasi kesepakatan awal Planned Contribution % seluruh anggota tim dengan validasi total tepat 100,00% (BRULE-PRJ-002, BR-016).
func (s *Service) FinalizePlannedContribution(ctx context.Context, projectID uuid.UUID, req *FinalizePlannedContributionRequest) error {
	var totalPct float64
	for _, c := range req.Contributions {
		totalPct += c.PlannedContributionPct
	}

	// Validasi Invariant: Total Planned Contribution wajib tepat 100.00%
	if math.Abs(totalPct-100.00) > 0.01 {
		return fmt.Errorf("total kesepakatan Planned Contribution seluruh anggota tim harus tepat 100.00%% (jumlah saat ini: %.2f%%)", totalPct)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi penetapan kontribusi: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, c := range req.Contributions {
		if err := s.repo.UpdatePlannedContributionTx(ctx, tx, projectID, c.UserID, c.PlannedContributionPct); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("gagal melakukan commit penetapan planned contribution: %w", err)
	}

	return nil
}

// Menghitung persentase Actual Contribution (Layer 2) berbasis tugas COMPLETED dan bobot kesulitan riil.
func (s *Service) CalculateActualContribution(ctx context.Context, projectID uuid.UUID) ([]ProjectTeamMember, error) {
	tasks, err := s.repo.GetCompletedTasksByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	members, err := s.repo.ListTeamMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, errors.New("tidak ada anggota tim pada proyek ini")
	}

	// Hitung skor tugas berbobot per anggota tim: difficulty_weight * estimated_hours
	memberScores := make(map[uuid.UUID]float64)
	var totalScore float64

	for _, t := range tasks {
		if t.AssigneeID != nil {
			weight := float64(t.DifficultyWeight)
			if weight <= 0 {
				weight = 1
			}
			hours := t.EstimatedHours
			if hours <= 0 {
				hours = 1
			}
			taskScore := weight * hours
			memberScores[*t.AssigneeID] += taskScore
			totalScore += taskScore
		}
	}

	// Hitung persentase Actual % untuk setiap anggota
	for i := range members {
		m := &members[i]
		actualPct := 0.00
		if totalScore > 0 {
			rawScore := memberScores[m.UserID]
			actualPct = math.Round((rawScore/totalScore)*10000) / 100
		} else {
			// Jika belum ada tugas selesai, default ke Planned %
			actualPct = m.PlannedContributionPct
		}

		_ = s.repo.UpdateActualContribution(ctx, projectID, m.UserID, actualPct)
		m.ActualContributionPct = &actualPct
	}

	return members, nil
}

// Mengesahkan dan mengunci Final Contribution % (Layer 3) oleh Supervisor dengan validasi total tepat 100,00% (BR-016).
func (s *Service) FinalizeContribution(ctx context.Context, projectID uuid.UUID, req *FinalizeContributionRequest, supervisorID uuid.UUID) error {
	var totalPct float64
	for _, c := range req.FinalContributions {
		totalPct += c.FinalContributionPct
	}

	// Penegakan aturan ketat total Final % = 100.00%
	if math.Abs(totalPct-100.00) > 0.01 {
		return fmt.Errorf("pengesahan Final Contribution gagal: total persentase harus tepat 100.00%% (jumlah saat ini: %.2f%%)", totalPct)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi pengesahan kontribusi final: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, c := range req.FinalContributions {
		if err := s.repo.LockFinalContributionTx(ctx, tx, projectID, c.UserID, c.FinalContributionPct); err != nil {
			return err
		}
	}

	// Ubah status proyek menjadi COMPLETED jika belum
	_, _ = tx.Exec(ctx, "UPDATE projects SET status = 'COMPLETED', updated_at = CURRENT_TIMESTAMP WHERE id = $1", projectID)

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("gagal melakukan commit pengesahan kontribusi final: %w", err)
	}

	// Terbitkan event domain untuk konsumsi downstream modul Finance (bounty distribution)
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "project.contribution_finalized",
			AggregateID: projectID.String(),
			Payload: map[string]interface{}{
				"project_id":       projectID.String(),
				"supervisor_id":    supervisorID.String(),
				"evaluation_notes": req.EvaluationNotes,
			},
		})
	}

	return nil
}

// ============================================================================
// 4. MILESTONES (FR-020, T-048)
// ============================================================================

// Membuat milestone target fase proyek baru dengan verifikasi batas total bobot 100%.
func (s *Service) CreateMilestone(ctx context.Context, projectID uuid.UUID, req *CreateMilestoneRequest) (*Milestone, error) {
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		deadline, err = time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			return nil, fmt.Errorf("format tenggat waktu milestone tidak valid: %w", err)
		}
	}

	currentWeight, err := s.repo.GetTotalMilestoneWeight(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if currentWeight+req.WeightPct > 100.01 {
		return nil, fmt.Errorf("total bobot milestone melebihi batas 100%% (bobot saat ini: %.2f%%, tambahan: %.2f%%)", currentWeight, req.WeightPct)
	}

	m := &Milestone{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		Deadline:    deadline,
		WeightPct:   req.WeightPct,
		Status:      MilestoneStatusPending,
	}

	if err := s.repo.CreateMilestone(ctx, m); err != nil {
		return nil, err
	}

	return m, nil
}

// ============================================================================
// 5. TASKS KANBAN STATE MACHINE (FR-021, T-049)
// ============================================================================

var validTaskTransitions = map[string]map[string]bool{
	TaskStatusTodo: {
		TaskStatusInProgress: true,
		TaskStatusBlocked:    true,
		TaskStatusCancelled:  true,
	},
	TaskStatusInProgress: {
		TaskStatusInReview:  true,
		TaskStatusBlocked:   true,
		TaskStatusTodo:      true,
		TaskStatusCancelled: true,
	},
	TaskStatusInReview: {
		TaskStatusCompleted:  true, // Hanya boleh melalui approval laporan tugas (BRULE-PRJ-003)
		TaskStatusInProgress: true, // Saat revisi diminta
	},
	TaskStatusBlocked: {
		TaskStatusTodo:       true,
		TaskStatusInProgress: true,
	},
	TaskStatusCompleted: {},
	TaskStatusCancelled: {},
}

// Membuat tugas kanban baru pada proyek.
func (s *Service) CreateTask(ctx context.Context, projectID uuid.UUID, req *CreateTaskRequest) (*Task, error) {
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		deadline, err = time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			return nil, fmt.Errorf("format tenggat waktu tugas tidak valid: %w", err)
		}
	}

	// Validasi penugasan anggota tim
	if req.AssigneeID != nil {
		isMember, _ := s.repo.GetTeamMember(ctx, projectID, *req.AssigneeID)
		if isMember == nil {
			return nil, errors.New("pengguna yang ditugaskan harus menjadi anggota tim proyek")
		}
	}

	task := &Task{
		ID:               uuid.New(),
		ProjectID:        projectID,
		MilestoneID:      req.MilestoneID,
		AssigneeID:       req.AssigneeID,
		Title:            req.Title,
		Description:      req.Description,
		EstimatedHours:   req.EstimatedHours,
		DifficultyWeight: req.DifficultyWeight,
		Priority:         strings.ToUpper(req.Priority),
		Deadline:         deadline,
		Status:           TaskStatusTodo,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// Mengubah status tugas kanban sesuai dengan aturan matriks transisi mesin status.
func (s *Service) ChangeTaskStatus(ctx context.Context, taskID uuid.UUID, targetStatus string) (*Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("tugas tidak ditemukan")
	}

	target := strings.ToUpper(targetStatus)
	if task.Status == target {
		return task, nil
	}

	// Validasi transisi status
	allowedTargets, exists := validTaskTransitions[task.Status]
	if !exists || !allowedTargets[target] {
		return nil, fmt.Errorf("transisi status tugas dari '%s' ke '%s' tidak diizinkan oleh mesin status kanban", task.Status, target)
	}

	// Tugas hanya boleh COMPLETED jika work report disetujui reviewer (BRULE-PRJ-003)
	if target == TaskStatusCompleted {
		return nil, errors.New("tugas tidak dapat langsung diselesaikan tanpa melalui penyerahan laporan dan verifikasi bukti kerja (BRULE-PRJ-003)")
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := s.repo.UpdateTaskStatusTx(ctx, tx, taskID, target); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	task.Status = target
	return task, nil
}

// ============================================================================
// 6. WORK REPORTS & EVIDENCE (FR-022, FR-023, T-050, T-051)
// ============================================================================

// Menyerahkan laporan pekerjaan tugas beserta bukti deliverable fisik di R2 atau tautan commit Git (DATA-004).
func (s *Service) SubmitWorkReport(ctx context.Context, taskID, userID uuid.UUID, req *SubmitWorkReportRequest) (*WorkReport, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("tugas tidak ditemukan")
	}

	// Validasi kepemilikan penugasan
	if task.AssigneeID != nil && *task.AssigneeID != userID {
		isMember, _ := s.repo.GetTeamMember(ctx, task.ProjectID, userID)
		if isMember == nil {
			return nil, errors.New("anda tidak memiliki hak akses untuk menyerahkan laporan pada tugas ini")
		}
	}

	evType := strings.ToUpper(req.EvidenceType)
	// Validasi format URL jika tipe Git Commit atau URL
	if evType == EvidenceTypeGitCommit || evType == EvidenceTypeURL {
		parsedURL, errURL := url.ParseRequestURI(req.EvidenceURLOrKey)
		if errURL != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return nil, fmt.Errorf("tautan bukti deliverable '%s' tidak valid, harus berupa URL web yang sah", req.EvidenceURLOrKey)
		}
	}

	rep := &WorkReport{
		ID:                 uuid.New(),
		TaskID:             taskID,
		UserID:             userID,
		Date:               time.Now().In(WIB),
		ProgressPercentage: req.ProgressPercentage,
		WhatIDid:           req.WhatIDid,
		EvidenceType:       evType,
		EvidenceURLOrKey:   req.EvidenceURLOrKey,
		Problems:           req.Problems,
		NextActions:        req.NextActions,
		Status:             ReportStatusSubmitted,
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi penyerahan tugas: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.CreateWorkReportTx(ctx, tx, rep); err != nil {
		return nil, err
	}

	// Alihkan status tugas menjadi IN_REVIEW secara atomik
	if err := s.repo.UpdateTaskStatusTx(ctx, tx, taskID, TaskStatusInReview); err != nil {
		return nil, err
	}

	// Simpan artefak bukti ke tabel evidence dalam transaksi
	evidence := &Evidence{
		ID:           uuid.New(),
		SubmissionID: rep.ID,
		EvidenceType: evType,
		ExternalURL:  &req.EvidenceURLOrKey,
		Description:  &req.WhatIDid,
	}
	if err := s.repo.CreateEvidenceTx(ctx, tx, evidence); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit penyerahan tugas: %w", err)
	}

	return rep, nil
}

// Meninjau dan menyetujui atau meminta revisi atas laporan pekerjaan tugas oleh Reviewer (BRULE-PRJ-003).
func (s *Service) ReviewWorkReport(ctx context.Context, reportID uuid.UUID, action string) (*WorkReport, error) {
	rep, err := s.repo.GetWorkReportByID(ctx, reportID)
	if err != nil {
		return nil, err
	}
	if rep == nil {
		return nil, errors.New("laporan pekerjaan tidak ditemukan")
	}

	upperAction := strings.ToUpper(action)
	if upperAction != "APPROVE" && upperAction != "REVISION_REQUIRED" {
		return nil, errors.New("aksi ulasan tidak valid, diharapkan 'APPROVE' atau 'REVISION_REQUIRED'")
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi ulasan laporan: %w", err)
	}
	defer tx.Rollback(ctx)

	if upperAction == "APPROVE" {
		if err := s.repo.UpdateWorkReportStatusTx(ctx, tx, reportID, ReportStatusApproved); err != nil {
			return nil, err
		}
		// Selesaikan tugas menjadi COMPLETED secara sah setelah approval bukti (BRULE-PRJ-003)
		if err := s.repo.UpdateTaskStatusTx(ctx, tx, rep.TaskID, TaskStatusCompleted); err != nil {
			return nil, err
		}
		rep.Status = ReportStatusApproved
	} else {
		if err := s.repo.UpdateWorkReportStatusTx(ctx, tx, reportID, ReportStatusRevisionRequired); err != nil {
			return nil, err
		}
		// Kembalikan tugas ke IN_PROGRESS untuk direvisi pelaksana
		if err := s.repo.UpdateTaskStatusTx(ctx, tx, rep.TaskID, TaskStatusInProgress); err != nil {
			return nil, err
		}
		rep.Status = ReportStatusRevisionRequired
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit ulasan laporan tugas: %w", err)
	}

	return rep, nil
}

// ============================================================================
// 7. SKILL MATCHING ENGINE (FR-018, T-053 - ADVISORY ONLY)
// ============================================================================

// Menghitung persentase kesesuaian keahlian kandidat pelamar terhadap kebutuhan proyek murni sebagai informasi penasihat (BR-015, BRULE-PRJ-005).
func (s *Service) CalculateSkillMatchPercentage(ctx context.Context, requiredSkills []uuid.UUID, userID uuid.UUID) float64 {
	if len(requiredSkills) == 0 {
		return 100.00
	}

	userSkills, err := s.getUserSkills(ctx, userID)
	if err != nil || len(userSkills) == 0 {
		return 0.00
	}

	matchedCount := 0
	for _, reqSkill := range requiredSkills {
		if userSkills[reqSkill] {
			matchedCount++
		}
	}

	score := (float64(matchedCount) / float64(len(requiredSkills))) * 100.00
	return math.Round(score*100) / 100
}

// Helper untuk mengambil set keahlian yang dimiliki oleh pengguna
func (s *Service) getUserSkills(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	query := `SELECT skill_id FROM user_skills WHERE user_id = $1`
	rows, err := s.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := make(map[uuid.UUID]bool)
	for rows.Next() {
		var sID uuid.UUID
		if err := rows.Scan(&sID); err == nil {
			skills[sID] = true
		}
	}
	return skills, nil
}
