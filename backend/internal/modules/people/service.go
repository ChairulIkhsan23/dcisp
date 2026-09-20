package people

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dcisp/backend/internal/database"
	"dcisp/backend/internal/integrations/apiindonesia"
	"dcisp/backend/internal/modules/system"
	"dcisp/backend/internal/shared/eventbus"
	"github.com/google/uuid"
)

type Service struct {
	repo           *Repository
	db             *database.PostgresDB
	auditService   *system.AuditService
	eventBus       *eventbus.EventBus
	externalClient *apiindonesia.Client
	redisClient    *database.RedisClient
}

// Menginisialisasi instance baru service people and lifecycle management.
func NewService(repo *Repository, db *database.PostgresDB, audit *system.AuditService, bus *eventbus.EventBus) *Service {
	return &Service{
		repo:         repo,
		db:           db,
		auditService: audit,
		eventBus:     bus,
	}
}

// Menetapkan klien eksternal API Indonesia dan Redis cache untuk pencarian institusi.
func (s *Service) SetExternalDependencies(client *apiindonesia.Client, rdb *database.RedisClient) {
	s.externalClient = client
	s.redisClient = rdb
}

// ============================================================================
// 1. INSTITUTIONS
// ============================================================================

// Membuat data institusi pendidikan mitra baru.
func (s *Service) CreateInstitution(ctx context.Context, req *CreateInstitutionRequest) (*Institution, error) {
	inst := &Institution{
		ID:            uuid.New(),
		Name:          req.Name,
		Address:       req.Address,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
	}
	if err := s.repo.CreateInstitution(ctx, inst); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "CREATE_INSTITUTION",
			ResourceType: "INSTITUTION",
			ResourceID:   inst.ID.String(),
			NewState: map[string]interface{}{
				"name": inst.Name,
			},
		})
	}

	return inst, nil
}

// Mengambil detail informasi institusi pendidikan berdasarkan ID.
func (s *Service) GetInstitutionByID(ctx context.Context, id uuid.UUID) (*Institution, error) {
	inst, err := s.repo.GetInstitutionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, errors.New("institusi tidak ditemukan")
	}
	return inst, nil
}

// Mengambil seluruh daftar institusi pendidikan mitra.
func (s *Service) ListInstitutions(ctx context.Context) ([]Institution, error) {
	return s.repo.ListInstitutions(ctx)
}

// Memperbarui data institusi pendidikan mitra.
func (s *Service) UpdateInstitution(ctx context.Context, id uuid.UUID, req *UpdateInstitutionRequest) (*Institution, error) {
	inst, err := s.repo.GetInstitutionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, errors.New("institusi tidak ditemukan")
	}

	if req.Name != "" {
		inst.Name = req.Name
	}
	if req.Address != nil {
		inst.Address = req.Address
	}
	if req.ContactPerson != nil {
		inst.ContactPerson = req.ContactPerson
	}
	if req.Email != nil {
		inst.Email = req.Email
	}
	if req.Phone != nil {
		inst.Phone = req.Phone
	}

	if err := s.repo.UpdateInstitution(ctx, inst); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "UPDATE_INSTITUTION",
			ResourceType: "INSTITUTION",
			ResourceID:   id.String(),
			NewState: map[string]interface{}{
				"name": inst.Name,
			},
		})
	}

	return inst, nil
}

// Menghapus institusi pendidikan dengan perlindungan pencegahan jika masih digunakan (BRULE-PEO-005).
func (s *Service) DeleteInstitution(ctx context.Context, id uuid.UUID) error {
	count, err := s.repo.CountInternsByInstitution(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("institusi tidak dapat dihapus karena masih memiliki %d peserta magang terhubung (BRULE-PEO-005)", count)
	}

	if err := s.repo.DeleteInstitution(ctx, id); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "DELETE_INSTITUTION",
			ResourceType: "INSTITUTION",
			ResourceID:   id.String(),
		})
	}

	return nil
}

// ============================================================================
// 2. BATCHES & BATCH FUNDS
// ============================================================================

// Membuat kohort batch baru dan menginisialisasi Batch Fund secara otomatis dalam satu transaksi (FR-004, BRULE-PEO-004).
func (s *Service) CreateBatch(ctx context.Context, req *CreateBatchRequest) (*Batch, *BatchFund, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, nil, fmt.Errorf("format tanggal mulai tidak valid (YYYY-MM-DD): %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, nil, fmt.Errorf("format tanggal selesai tidak valid (YYYY-MM-DD): %w", err)
	}
	if endDate.Before(startDate) {
		return nil, nil, errors.New("tanggal selesai tidak boleh mendahului tanggal mulai")
	}

	status := BatchStatusDraft
	if req.Status != "" {
		status = req.Status
	}

	batch := &Batch{
		ID:        uuid.New(),
		BatchCode: req.BatchCode,
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
		Quota:     req.Quota,
		Status:    status,
	}

	// Transaksi atomik: Buat batch + Buat Batch Fund
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal memulai transaksi database: %w", err)
	}
	defer tx.Rollback(ctx)

	fund, err := s.repo.CreateBatchWithFundTx(ctx, tx, batch)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("gagal melakukan commit transaksi pembuatan batch: %w", err)
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "CREATE_BATCH",
			ResourceType: "BATCH",
			ResourceID:   batch.ID.String(),
			NewState: map[string]interface{}{
				"batch_code": batch.BatchCode,
				"quota":      batch.Quota,
				"fund_id":    fund.ID.String(),
			},
		})
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "batch.created",
			AggregateID: batch.ID.String(),
			Payload: map[string]interface{}{
				"batch_code": batch.BatchCode,
				"quota":      batch.Quota,
			},
		})
	}

	return batch, fund, nil
}

// Mengambil informasi detail kohort batch beserta kas bersamanya berdasarkan ID.
func (s *Service) GetBatchByID(ctx context.Context, id uuid.UUID) (*Batch, *BatchFund, error) {
	batch, err := s.repo.GetBatchByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if batch == nil {
		return nil, nil, errors.New("batch tidak ditemukan")
	}

	fund, _ := s.repo.GetBatchFundByBatchID(ctx, id)
	return batch, fund, nil
}

// Mengambil daftar seluruh kohort batch dengan opsi filter status.
func (s *Service) ListBatches(ctx context.Context, status string) ([]Batch, error) {
	return s.repo.ListBatches(ctx, status)
}

// Memperbarui informasi kohort batch.
func (s *Service) UpdateBatch(ctx context.Context, id uuid.UUID, req *UpdateBatchRequest) (*Batch, error) {
	batch, err := s.repo.GetBatchByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, errors.New("batch tidak ditemukan")
	}

	if req.Name != "" {
		batch.Name = req.Name
	}
	if req.StartDate != "" {
		st, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("format tanggal mulai tidak valid: %w", err)
		}
		batch.StartDate = st
	}
	if req.EndDate != "" {
		et, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("format tanggal selesai tidak valid: %w", err)
		}
		batch.EndDate = et
	}
	if batch.EndDate.Before(batch.StartDate) {
		return nil, errors.New("tanggal selesai tidak boleh mendahului tanggal mulai")
	}
	if req.Quota != nil {
		if *req.Quota < 1 {
			return nil, errors.New("kuota batch minimal 1")
		}
		batch.Quota = *req.Quota
	}
	if req.Status != "" {
		batch.Status = req.Status
	}

	if err := s.repo.UpdateBatch(ctx, batch); err != nil {
		return nil, err
	}

	return batch, nil
}

// ============================================================================
// 3. INTERN LIFECYCLE & STATE MACHINE (FR-002)
// ============================================================================

// Matriks transisi status resmi siklus hidup peserta magang (FR-002)
var validTransitions = map[string]map[string]bool{
	InternStatusApplicant: {
		InternStatusOnboarding: true,
		InternStatusTerminated: true,
	},
	InternStatusOnboarding: {
		InternStatusActive:     true,
		InternStatusTerminated: true,
	},
	InternStatusActive: {
		InternStatusOnLeave:    true,
		InternStatusSuspended:  true,
		InternStatusGraduated:  true,
		InternStatusTerminated: true,
	},
	InternStatusOnLeave: {
		InternStatusActive:     true,
		InternStatusTerminated: true,
	},
	InternStatusSuspended: {
		InternStatusActive:     true,
		InternStatusTerminated: true,
	},
	InternStatusGraduated:  {}, // State akhir magang (beralih ke Alumni)
	InternStatusTerminated: {}, // State akhir diskualifikasi
}

// Mendaftarkan peserta magang baru dengan status awal APPLICANT.
func (s *Service) RegisterIntern(ctx context.Context, req *RegisterInternRequest) (*Intern, error) {
	// Pastikan batch ada dan berstatus aktif/draft
	batch, err := s.repo.GetBatchByID(ctx, req.BatchID)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, errors.New("batch tidak ditemukan")
	}

	joinDate, err := time.Parse("2006-01-02", req.JoinDate)
	if err != nil {
		return nil, fmt.Errorf("format tanggal mulai magang tidak valid: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("format tanggal selesai magang tidak valid: %w", err)
	}

	intern := &Intern{
		ID:            uuid.New(),
		UserID:        req.UserID,
		BatchID:       req.BatchID,
		InstitutionID: req.InstitutionID,
		IDNumber:      req.IDNumber,
		Status:        InternStatusApplicant,
		JoinDate:      joinDate,
		EndDate:       endDate,
		InternshipXP:  0,
	}

	if err := s.repo.CreateIntern(ctx, intern); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "REGISTER_INTERN",
			ResourceType: "INTERN",
			ResourceID:   intern.ID.String(),
			NewState: map[string]interface{}{
				"user_id":  intern.UserID.String(),
				"batch_id": intern.BatchID.String(),
				"status":   intern.Status,
			},
		})
	}

	return intern, nil
}

// Memproses transisi status siklus hidup peserta magang berdasarkan aturan state machine yang sah.
func (s *Service) ChangeInternStatus(ctx context.Context, internID uuid.UUID, newStatus string, reason string) (*Intern, error) {
	intern, err := s.repo.GetInternByID(ctx, internID)
	if err != nil {
		return nil, err
	}
	if intern == nil {
		return nil, errors.New("peserta magang tidak ditemukan")
	}

	if intern.Status == newStatus {
		return nil, fmt.Errorf("status peserta magang sudah '%s'", newStatus)
	}

	// Validasi transisi status
	allowedTargets, exists := validTransitions[intern.Status]
	if !exists || !allowedTargets[newStatus] {
		return nil, fmt.Errorf("transisi status dari '%s' ke '%s' tidak diizinkan oleh mesin status (FR-002)", intern.Status, newStatus)
	}

	oldStatus := intern.Status

	// Transaksi untuk menangani side effects
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi transisi status: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Eksekusi transisi penerimaan (dari APPLICANT ke ONBOARDING atau ACTIVE): Periksa kuota batch
	if oldStatus == InternStatusApplicant && (newStatus == InternStatusOnboarding || newStatus == InternStatusActive) {
		lockedBatch, err := s.repo.LockBatchForUpdateTx(ctx, tx, intern.BatchID)
		if err != nil {
			return nil, err
		}
		count, err := s.repo.CountInternsInBatch(ctx, intern.BatchID)
		if err != nil {
			return nil, err
		}
		if count >= lockedBatch.Quota {
			return nil, fmt.Errorf("kuota batch '%s' telah penuh (%d/%d)", lockedBatch.Name, count, lockedBatch.Quota)
		}
	}

	// 2. Eksekusi transisi ke GRADUATED: Buat profil Alumni secara otomatis & alihkan peran (FR-003, BRULE-PEO-001)
	if newStatus == InternStatusGraduated {
		alumni := &Alumni{
			ID:              uuid.New(),
			UserID:          intern.UserID,
			BatchID:         intern.BatchID,
			GraduationDate:  time.Now().UTC(),
			AlumniXP:        0,
			IsPublicProfile: true,
		}
		if err := s.repo.CreateAlumniTx(ctx, tx, alumni); err != nil {
			return nil, fmt.Errorf("gagal membuat profil alumni saat kelulusan: %w", err)
		}

		// Alihkan peran sistem menjadi ALUMNI
		if err := s.repo.SwitchUserRoleToAlumniTx(ctx, tx, intern.UserID); err != nil {
			return nil, fmt.Errorf("gagal mengalihkan peran pengguna ke ALUMNI: %w", err)
		}
	}

	// 3. Update status intern
	if err := s.repo.UpdateInternStatusTx(ctx, tx, intern.ID, newStatus); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit transaksi status: %w", err)
	}

	intern.Status = newStatus

	// Audit Logging
	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "INTERN_STATUS_CHANGED",
			ResourceType: "INTERN",
			ResourceID:   intern.ID.String(),
			OldState: map[string]interface{}{
				"status": oldStatus,
			},
			NewState: map[string]interface{}{
				"status": newStatus,
				"reason": reason,
			},
		})
	}

	// Domain Event
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "intern.status_changed",
			AggregateID: intern.ID.String(),
			Payload: map[string]interface{}{
				"intern_id":  intern.ID.String(),
				"user_id":    intern.UserID.String(),
				"old_status": oldStatus,
				"new_status": newStatus,
			},
		})

		if newStatus == InternStatusGraduated {
			_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
				Type:        "intern.graduated",
				AggregateID: intern.UserID.String(),
				Payload: map[string]interface{}{
					"user_id":  intern.UserID.String(),
					"batch_id": intern.BatchID.String(),
				},
			})
		}
	}

	return intern, nil
}

// Mengambil data lengkap peserta magang berdasarkan ID profil.
func (s *Service) GetInternByID(ctx context.Context, id uuid.UUID) (*Intern, error) {
	intern, err := s.repo.GetInternByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if intern == nil {
		return nil, errors.New("peserta magang tidak ditemukan")
	}
	return intern, nil
}

// Mengambil data lengkap peserta magang berdasarkan user ID akun.
func (s *Service) GetInternByUserID(ctx context.Context, userID uuid.UUID) (*Intern, error) {
	intern, err := s.repo.GetInternByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if intern == nil {
		return nil, errors.New("peserta magang tidak ditemukan")
	}
	return intern, nil
}

// ============================================================================
// 4. USER MANAGEMENT (SEARCH, MENTOR ASSIGNMENT, BATCH PLOTTING - T-032)
// ============================================================================

// Mencari peserta magang berdasarkan nama, email, NIM, status, institusi, dan batch dengan paginasi.
func (s *Service) SearchInterns(ctx context.Context, search string, batchID, institutionID *uuid.UUID, status string, page, limit int) ([]Intern, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.repo.ListInterns(ctx, search, batchID, institutionID, status, limit, offset)
}

// Menetapkan pembimbing (mentor) bagi peserta magang tertentu.
func (s *Service) AssignMentor(ctx context.Context, internID, mentorID uuid.UUID) (*Intern, error) {
	intern, err := s.repo.GetInternByID(ctx, internID)
	if err != nil {
		return nil, err
	}
	if intern == nil {
		return nil, errors.New("peserta magang tidak ditemukan")
	}

	if err := s.repo.UpdateInternMentor(ctx, internID, &mentorID); err != nil {
		return nil, err
	}

	intern.MentorID = &mentorID

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "ASSIGN_MENTOR",
			ResourceType: "INTERN",
			ResourceID:   internID.String(),
			NewState: map[string]interface{}{
				"mentor_id": mentorID.String(),
			},
		})
	}

	return intern, nil
}

// Memindahkan atau menetapkan peserta magang ke kohort batch baru dengan validasi kuota aman konkurensi.
func (s *Service) PlotBatch(ctx context.Context, internID, newBatchID uuid.UUID) (*Intern, error) {
	intern, err := s.repo.GetInternByID(ctx, internID)
	if err != nil {
		return nil, err
	}
	if intern == nil {
		return nil, errors.New("peserta magang tidak ditemukan")
	}

	if intern.BatchID == newBatchID {
		return intern, nil
	}

	// Transaksi aman konkurensi dengan penguncian batch
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi plotting batch: %w", err)
	}
	defer tx.Rollback(ctx)

	targetBatch, err := s.repo.LockBatchForUpdateTx(ctx, tx, newBatchID)
	if err != nil {
		return nil, err
	}
	if targetBatch == nil {
		return nil, errors.New("batch target tidak ditemukan")
	}

	currentCount, err := s.repo.CountInternsInBatch(ctx, newBatchID)
	if err != nil {
		return nil, err
	}
	if currentCount >= targetBatch.Quota {
		return nil, fmt.Errorf("kuota batch '%s' telah terpenuhi (%d/%d)", targetBatch.Name, currentCount, targetBatch.Quota)
	}

	if err := s.repo.UpdateInternBatchTx(ctx, tx, internID, newBatchID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit plotting batch: %w", err)
	}

	intern.BatchID = newBatchID
	intern.BatchName = targetBatch.Name

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "PLOT_BATCH",
			ResourceType: "INTERN",
			ResourceID:   internID.String(),
			NewState: map[string]interface{}{
				"batch_id": newBatchID.String(),
			},
		})
	}

	return intern, nil
}

// ============================================================================
// 5. ALUMNI (FR-003, BR-003)
// ============================================================================

// Mengambil profil alumni berdasarkan user ID akun dengan retensi data permanen.
func (s *Service) GetAlumniByUserID(ctx context.Context, userID uuid.UUID) (*Alumni, error) {
	alumni, err := s.repo.GetAlumniByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if alumni == nil {
		return nil, errors.New("data alumni tidak ditemukan")
	}
	return alumni, nil
}

// Mengambil daftar alumni terdaftar dengan paginasi.
func (s *Service) ListAlumni(ctx context.Context, page, limit int) ([]Alumni, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListAlumni(ctx, limit, offset)
}

// ============================================================================
// 6. SKILLS & USER SKILLS (FR-006)
// ============================================================================

// Membuat master keahlian teknis baru ke dalam katalog skills.
func (s *Service) CreateSkill(ctx context.Context, req *CreateSkillRequest) (*Skill, error) {
	skill := &Skill{
		ID:          uuid.New(),
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
	}
	if err := s.repo.CreateSkill(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

// Mengambil seluruh master katalog skill dengan filter kategori opsional.
func (s *Service) ListSkills(ctx context.Context, category string) ([]Skill, error) {
	return s.repo.ListSkills(ctx, category)
}

// Menetapkan atau memperbarui tingkat kemahiran keahlian peserta (skala 1 s/d 5).
func (s *Service) AssignUserSkill(ctx context.Context, userID uuid.UUID, req *AssignUserSkillRequest) (*UserSkill, error) {
	if req.ProficiencyLevel < 1 || req.ProficiencyLevel > 5 {
		return nil, errors.New("tingkat kemahiran harus bernilai antara 1 sampai 5")
	}

	us := &UserSkill{
		ID:               uuid.New(),
		UserID:           userID,
		SkillID:          req.SkillID,
		ProficiencyLevel: req.ProficiencyLevel,
	}

	if err := s.repo.UpsertUserSkill(ctx, us); err != nil {
		return nil, err
	}

	return us, nil
}

// Mengambil daftar seluruh keahlian yang dimiliki oleh pengguna tertentu.
func (s *Service) GetUserSkills(ctx context.Context, userID uuid.UUID) ([]UserSkill, error) {
	return s.repo.GetUserSkills(ctx, userID)
}

// Menghapus rekaman keahlian yang terhubung dengan pengguna.
func (s *Service) DeleteUserSkill(ctx context.Context, userID, skillID uuid.UUID) error {
	return s.repo.DeleteUserSkill(ctx, userID, skillID)
}
