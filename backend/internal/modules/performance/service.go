package performance

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/system"
	"dcisp/backend/internal/shared/eventbus"
	"github.com/google/uuid"
)

type Service struct {
	repo         *Repository
	db           *database.PostgresDB
	auditService *system.AuditService
	eventBus     *eventbus.EventBus
}

// Menginisialisasi instance baru service performance evaluation and gamification management.
func NewService(repo *Repository, db *database.PostgresDB, audit *system.AuditService, bus *eventbus.EventBus) *Service {
	s := &Service{
		repo:         repo,
		db:           db,
		auditService: audit,
		eventBus:     bus,
	}

	// Daftarkan listener event bus asinkron untuk konsumsi event kehadiran dan tugas (T-055, T-057)
	if bus != nil {
		s.registerEventListeners()
	}

	return s
}

// Mendaftarkan fungsi pendengar domain event untuk pemrosesan reward dan penalti XP otomatis.
func (s *Service) registerEventListeners() {
	// 1. Konsumsi event presensi kehadiran (T-055 - FR-025)
	s.eventBus.Subscribe("attendance.scanned", func(ctx context.Context, event eventbus.DomainEvent) error {
		uIDStr, ok := event.Payload["user_id"].(string)
		if !ok {
			return nil
		}
		uID, err := uuid.Parse(uIDStr)
		if err != nil {
			return nil
		}

		eventType, _ := event.Payload["event_type"].(string)
		isLate, _ := event.Payload["is_late"].(bool)
		lateTierFloat, _ := event.Payload["late_tier"].(float64)
		lateTier := int(lateTierFloat)
		ts, _ := event.Payload["timestamp"].(string)

		if eventType == "CHECK_IN" {
			ref := fmt.Sprintf("attendance:check_in:%s:%s", uIDStr, ts)
			if !isLate {
				_, _ = s.RecordXPMutation(ctx, &RecordXPMutationRequest{
					UserID:         uID,
					EventTrigger:   "CHECK_IN_ON_TIME",
					ReferenceEvent: ref,
				})
			} else {
				trigger := "LATE_TIER_1"
				if lateTier == 2 {
					trigger = "LATE_TIER_2"
				} else if lateTier == 3 {
					trigger = "LATE_TIER_3"
				}
				_, _ = s.RecordXPMutation(ctx, &RecordXPMutationRequest{
					UserID:         uID,
					EventTrigger:   trigger,
					ReferenceEvent: ref,
				})
			}
		}
		return nil
	})

	// 2. Konsumsi event penyelesaian tugas proyek (T-057 - FR-025)
	s.eventBus.Subscribe("task.report_reviewed", func(ctx context.Context, event eventbus.DomainEvent) error {
		status, _ := event.Payload["status"].(string)
		if status != "APPROVED" {
			return nil
		}

		taskIDStr, _ := event.Payload["task_id"].(string)
		assigneeIDStr, _ := event.Payload["assignee_id"].(string)
		diffWeightFloat, _ := event.Payload["difficulty_weight"].(float64)
		diffWeight := int(diffWeightFloat)

		uID, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			return nil
		}

		trigger := "TASK_COMPLETED_LOW"
		if diffWeight == 3 {
			trigger = "TASK_COMPLETED_MED"
		} else if diffWeight >= 4 {
			trigger = "TASK_COMPLETED_HIGH"
		}

		ref := fmt.Sprintf("task:completion:%s", taskIDStr)
		_, _ = s.RecordXPMutation(ctx, &RecordXPMutationRequest{
			UserID:         uID,
			EventTrigger:   trigger,
			ReferenceEvent: ref,
		})

		return nil
	})
}

// ============================================================================
// 1. XP RULES ENGINE & IMMUTABLE LEDGER (FR-025, T-054)
// ============================================================================

// Mencatat mutasi poin pengalaman baru secara append-only ke dalam buku besar dengan pencegahan duplikasi idempoten (BR-004, BR-012, BRULE-PRF-001).
func (s *Service) RecordXPMutation(ctx context.Context, req *RecordXPMutationRequest) (*XPTransaction, error) {
	// 1. Idempotency Check: Pastikan event ini belum pernah menghasilkan mutasi XP sebelumnya
	alreadyProcessed, err := s.repo.HasProcessedEvent(ctx, req.UserID, req.ReferenceEvent)
	if err != nil {
		return nil, err
	}
	if alreadyProcessed {
		// Event duplikat diabaikan dengan aman tanpa menambahkan transaksi ganda
		lastBal, _ := s.repo.GetRunningBalance(ctx, req.UserID, XPSchemeInternship)
		return &XPTransaction{
			UserID:         req.UserID,
			RunningBalance: lastBal,
			ReferenceEvent: &req.ReferenceEvent,
		}, nil
	}

	// 2. Cari Aturan XP
	rule, err := s.repo.GetRuleByTrigger(ctx, req.EventTrigger)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, fmt.Errorf("aturan XP dengan trigger '%s' tidak ditemukan", req.EventTrigger)
	}

	points := rule.XPValue
	if req.PointsOverride != nil {
		points = *req.PointsOverride
	}

	scheme := rule.XPScheme

	// 3. Ambil Saldo Berjalan Terakhir untuk Skema Ini
	prevBalance, err := s.repo.GetRunningBalance(ctx, req.UserID, scheme)
	if err != nil {
		return nil, err
	}

	newBalance := prevBalance + points
	// Penegakan Batas Bawah: Saldo kumulatif Internship XP tidak boleh < 0 (BRULE-PRF-005)
	if scheme == XPSchemeInternship && newBalance < 0 {
		newBalance = 0
	}

	txRecord := &XPTransaction{
		ID:             uuid.New(),
		UserID:         req.UserID,
		XPRuleID:       &rule.ID,
		Scheme:         scheme,
		Points:         points,
		RunningBalance: newBalance,
		ReferenceEvent: &req.ReferenceEvent,
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi mutasi XP: %w", err)
	}
	defer tx.Rollback(ctx)

	// Simpan transaksi append-only (dilindungi trigger trg_immutable_xp_transactions)
	if err := s.repo.CreateTransactionTx(ctx, tx, txRecord); err != nil {
		return nil, err
	}

	// 4. Jika Skema Internship XP: Evaluasi Kenaikan Rank (T-056, NO AUTO-DEMOTION)
	promotedRankID := uuid.Nil
	if scheme == XPSchemeInternship {
		oldLevel, _ := s.repo.GetInternRankLevelTx(ctx, tx, req.UserID)
		eligibleRank, err := s.repo.GetRankByXP(ctx, newBalance)
		if err == nil && eligibleRank != nil {
			_ = s.repo.UpdateInternRankTx(ctx, tx, req.UserID, eligibleRank.ID, newBalance)
			if eligibleRank.LevelOrder > oldLevel {
				promotedRankID = eligibleRank.ID
			}
		}
	} else if scheme == XPSchemeAlumni {
		// Update saldo alumni_xp pada tabel alumni
		_, _ = tx.Exec(ctx, "UPDATE alumni SET alumni_xp = $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1", req.UserID, newBalance)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit transaksi XP: %w", err)
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "xp.mutated",
			AggregateID: req.UserID.String(),
			Payload: map[string]interface{}{
				"user_id":         req.UserID.String(),
				"scheme":          scheme,
				"points":          points,
				"running_balance": newBalance,
				"event_trigger":   req.EventTrigger,
			},
		})

		// Menerbitkan promosi rank agar penerbit reward menindaklanjutinya (F-EVT-02).
		if promotedRankID != uuid.Nil {
			_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
				Type:        "rank.promoted",
				AggregateID: req.UserID.String(),
				Payload: map[string]interface{}{
					"user_id": req.UserID.String(),
					"rank_id": promotedRankID.String(),
				},
			})
		}
	}

	return txRecord, nil
}

// Mengambil ringkasan saldo poin pengalaman pada 3 jalur skema terisolasi beserta tingkat rank saat ini (FR-025, FR-026).
func (s *Service) GetXPBalance(ctx context.Context, userID uuid.UUID) (*XPBalanceResponse, error) {
	internshipXP, _ := s.repo.GetRunningBalance(ctx, userID, XPSchemeInternship)
	projectXP, _ := s.repo.GetRunningBalance(ctx, userID, XPSchemeProject)
	alumniXP, _ := s.repo.GetRunningBalance(ctx, userID, XPSchemeAlumni)

	currentRank, _ := s.repo.GetRankByXP(ctx, internshipXP)
	var nextRank *Rank
	xpToNext := 0

	if currentRank != nil {
		nextRank, _ = s.repo.GetNextRank(ctx, currentRank.LevelOrder)
		if nextRank != nil {
			xpToNext = nextRank.MinXP - internshipXP
			if xpToNext < 0 {
				xpToNext = 0
			}
		}
	}

	return &XPBalanceResponse{
		UserID:               userID,
		InternshipXP:         internshipXP,
		ProjectXP:            projectXP,
		AlumniContributionXP: alumniXP,
		CurrentRank:          currentRank,
		NextRank:             nextRank,
		XPToNextRank:         xpToNext,
	}, nil
}

// ============================================================================
// 2. SUPERVISOR PERFORMANCE EVALUATION (FR-027, BR-017)
// ============================================================================

// Membuat dokumen evaluasi kinerja formal komposit dari supervisor (skala 0–100, terpisah dari XP).
func (s *Service) CreateEvaluation(ctx context.Context, evaluatorID uuid.UUID, req *CreateEvaluationRequest) (*PerformanceEvaluation, error) {
	// Formula Komposit Standar: Bobot berimbang 25% untuk 4 komponen rubrik
	composite := (req.AttendanceScore * 0.25) + (req.TaskDeliveryScore * 0.25) + (req.WorkQualityScore * 0.25) + (req.SupervisorRubricScore * 0.25)
	composite = math.Round(composite*100) / 100

	eval := &PerformanceEvaluation{
		ID:                        uuid.New(),
		UserID:                    req.UserID,
		EvaluatorID:               evaluatorID,
		BatchID:                   req.BatchID,
		PeriodType:                req.PeriodType,
		AttendanceScore:           req.AttendanceScore,
		TaskDeliveryScore:         req.TaskDeliveryScore,
		WorkQualityScore:          req.WorkQualityScore,
		SupervisorRubricScore:     req.SupervisorRubricScore,
		CompositePerformanceScore: composite,
		IsTopPerformer:            false,
		Notes:                     req.Notes,
	}

	if err := s.repo.CreateEvaluation(ctx, eval); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			UserID:       &evaluatorID,
			EventName:    "CREATE_PERFORMANCE_EVALUATION",
			ResourceType: "PERFORMANCE_EVALUATION",
			ResourceID:   eval.ID.String(),
			NewState: map[string]interface{}{
				"user_id":                     eval.UserID.String(),
				"composite_performance_score": composite,
				"period_type":                 eval.PeriodType,
			},
		})
	}

	return eval, nil
}

// Mengambil dokumen evaluasi kinerja formal berdasarkan ID unik.
func (s *Service) GetEvaluationByID(ctx context.Context, id uuid.UUID) (*PerformanceEvaluation, error) {
	eval, err := s.repo.GetEvaluationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if eval == nil {
		return nil, errors.New("evaluasi performa tidak ditemukan")
	}
	return eval, nil
}

// ============================================================================
// 3. TOP PERFORMER PER BATCH (FR-028, BR-018)
// ============================================================================

// Menentukan tepat SATU peserta terbaik (Top Performer / MVP) per kohort batch per periode evaluasi (BRULE-PRF-004).
func (s *Service) DetermineTopPerformer(ctx context.Context, batchID uuid.UUID, periodType string) (*TopPerformerResponse, error) {
	evals, err := s.repo.GetEvaluationsByBatchAndPeriod(ctx, batchID, periodType)
	if err != nil {
		return nil, err
	}
	if len(evals) == 0 {
		return nil, fmt.Errorf("tidak ditemukan evaluasi performa pada batch dan periode '%s'", periodType)
	}

	// Ambil kandidat dengan skor komposit tertinggi (urutan teratas hasil ORDER BY composite DESC, evaluated_at ASC)
	topCandidate := evals[0]

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi penentuan top performer: %w", err)
	}
	defer tx.Rollback(ctx)

	// Pastikan hanya tepat SATU pemenang per batch per periode: bersihkan status lama
	if err := s.repo.ClearTopPerformerInBatchTx(ctx, tx, batchID, periodType); err != nil {
		return nil, err
	}

	// Tetapkan kandidat terpilih sebagai Top Performer tunggal
	if err := s.repo.SetTopPerformerTx(ctx, tx, topCandidate.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit penetapan top performer: %w", err)
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "batch.top_performer_selected",
			AggregateID: batchID.String(),
			Payload: map[string]interface{}{
				"batch_id":                    batchID.String(),
				"winner_user_id":              topCandidate.UserID.String(),
				"composite_performance_score": topCandidate.CompositePerformanceScore,
			},
		})
	}

	return &TopPerformerResponse{
		BatchID:                   batchID,
		WinnerUserID:              topCandidate.UserID,
		WinnerFullName:            topCandidate.UserFullName,
		CompositePerformanceScore: topCandidate.CompositePerformanceScore,
		PeriodType:                periodType,
		EvaluationDate:            topCandidate.EvaluatedAt.Format(time.RFC3339),
	}, nil
}

// ============================================================================
// 4. ACHIEVEMENTS & BADGES (FR-029)
// ============================================================================

// Membuat master lencana prestasi baru ke dalam katalog achievements.
func (s *Service) CreateAchievement(ctx context.Context, req *CreateAchievementRequest) (*Achievement, error) {
	a := &Achievement{
		ID:           uuid.New(),
		Code:         req.Code,
		Title:        req.Title,
		Description:  req.Description,
		BadgeIconURL: req.BadgeIconURL,
		RewardXP:     req.RewardXP,
	}
	if err := s.repo.CreateAchievement(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Membuka lencana prestasi bagi pengguna secara idempoten dan memberikan bonus XP jika ada.
func (s *Service) UnlockAchievement(ctx context.Context, userID uuid.UUID, achievementCode string) (bool, error) {
	ach, err := s.repo.GetAchievementByCode(ctx, achievementCode)
	if err != nil {
		return false, err
	}
	if ach == nil {
		return false, fmt.Errorf("achievement dengan kode '%s' tidak ditemukan", achievementCode)
	}

	isNewUnlock, err := s.repo.UnlockAchievement(ctx, userID, ach.ID)
	if err != nil {
		return false, err
	}

	// Jika baru pertama kali dibuka dan memiliki bonus XP, berikan reward XP
	if isNewUnlock && ach.RewardXP > 0 {
		ref := fmt.Sprintf("achievement:unlock:%s", ach.Code)
		_, _ = s.RecordXPMutation(ctx, &RecordXPMutationRequest{
			UserID:         userID,
			EventTrigger:   "CHECK_IN_ON_TIME", // Fallback trigger
			PointsOverride: &ach.RewardXP,
			ReferenceEvent: ref,
		})
	}

	return isNewUnlock, nil
}

// Mengambil seluruh daftar lencana prestasi yang dimiliki oleh pengguna.
func (s *Service) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]UserAchievement, error) {
	return s.repo.GetUserAchievements(ctx, userID)
}

// ============================================================================
// 5. SKILL GROWTH MATRIX (FR-030, T-061)
// ============================================================================

// Mengagregasi matriks pertumbuhan kompetensi pengguna berdasarkan penyelesaian tugas proyek nyata.
func (s *Service) GetSkillGrowthMatrix(ctx context.Context, userID uuid.UUID) (*SkillGrowthMatrixResponse, error) {
	entries, err := s.repo.GetSkillGrowthMatrix(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &SkillGrowthMatrixResponse{
		UserID: userID,
		Skills: entries,
	}, nil
}
