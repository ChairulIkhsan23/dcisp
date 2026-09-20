package performance

import (
	"context"
	"errors"
	"fmt"

	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.PostgresDB
}

// Menginisialisasi instance baru repository performance and gamification management.
func NewRepository(db *database.PostgresDB) *Repository {
	return &Repository{db: db}
}

// ============================================================================
// 1. XP RULES & TRANSACTIONS (FR-025, BR-004, BR-012, BRULE-PRF-001)
// ============================================================================

// Mengambil master aturan XP berdasarkan kode pemicu peristiwanya.
func (r *Repository) GetRuleByTrigger(ctx context.Context, trigger string) (*XPRule, error) {
	query := `
		SELECT id, policy_id, event_trigger, xp_value, xp_scheme, description, created_at
		FROM xp_rules
		WHERE event_trigger = $1
	`
	var rule XPRule
	err := r.db.Pool.QueryRow(ctx, query, trigger).Scan(
		&rule.ID, &rule.PolicyID, &rule.EventTrigger, &rule.XPValue, &rule.XPScheme, &rule.Description, &rule.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari aturan XP berdasarkan pemicu: %w", err)
	}
	return &rule, nil
}

// Mengambil seluruh daftar master aturan XP yang terdaftar di sistem.
func (r *Repository) ListRules(ctx context.Context) ([]XPRule, error) {
	query := `
		SELECT id, policy_id, event_trigger, xp_value, xp_scheme, description, created_at
		FROM xp_rules
		ORDER BY xp_scheme ASC, event_trigger ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar aturan XP: %w", err)
	}
	defer rows.Close()

	var list []XPRule
	for rows.Next() {
		var rule XPRule
		if err := rows.Scan(&rule.ID, &rule.PolicyID, &rule.EventTrigger, &rule.XPValue, &rule.XPScheme, &rule.Description, &rule.CreatedAt); err == nil {
			list = append(list, rule)
		}
	}
	return list, nil
}

// Memeriksa apakah suatu peristiwa dengan referensi unik sudah pernah diproses untuk mencegah mutasi ganda (Idempotency).
func (r *Repository) HasProcessedEvent(ctx context.Context, userID uuid.UUID, referenceEvent string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM xp_transactions
		WHERE user_id = $1 AND reference_event = $2
	`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, userID, referenceEvent).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("gagal memeriksa idempotensi mutasi XP: %w", err)
	}
	return count > 0, nil
}

// Mengambil saldo berjalan (running balance) terakhir untuk skema XP tertentu.
func (r *Repository) GetRunningBalance(ctx context.Context, userID uuid.UUID, scheme string) (int, error) {
	query := `
		SELECT running_balance
		FROM xp_transactions
		WHERE user_id = $1 AND scheme = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var balance int
	err := r.db.Pool.QueryRow(ctx, query, userID, scheme).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("gagal mengambil saldo berjalan XP: %w", err)
	}
	return balance, nil
}

// Menyimpan rekaman mutasi XP secara append-only ke dalam buku besar xp_transactions dalam transaksi.
func (r *Repository) CreateTransactionTx(ctx context.Context, tx pgx.Tx, t *XPTransaction) error {
	query := `
		INSERT INTO xp_transactions (id, user_id, xp_rule_id, scheme, points, running_balance, reference_event, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
	`
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, query, t.ID, t.UserID, t.XPRuleID, t.Scheme, t.Points, t.RunningBalance, t.ReferenceEvent)
	if err != nil {
		return fmt.Errorf("gagal mencatat mutasi XP append-only: %w", err)
	}
	return nil
}

// Mengambil daftar riwayat transaksi XP pengguna dengan filter skema dan paginasi.
func (r *Repository) ListUserTransactions(ctx context.Context, userID uuid.UUID, scheme string, limit, offset int) ([]XPTransaction, int, error) {
	baseQuery := `
		FROM xp_transactions t
		LEFT JOIN xp_rules r ON r.id = t.xp_rule_id
		WHERE t.user_id = $1
	`
	args := []interface{}{userID}
	argIdx := 2

	if scheme != "" {
		baseQuery += fmt.Sprintf(" AND t.scheme = $%d", argIdx)
		args = append(args, scheme)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total transaksi XP: %w", err)
	}

	selectQuery := fmt.Sprintf(`
		SELECT t.id, t.user_id, t.xp_rule_id, t.scheme, t.points, t.running_balance, t.reference_event, t.created_at,
		       r.event_trigger
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil riwayat transaksi XP: %w", err)
	}
	defer rows.Close()

	var list []XPTransaction
	for rows.Next() {
		var txRecord XPTransaction
		if err := rows.Scan(
			&txRecord.ID, &txRecord.UserID, &txRecord.XPRuleID, &txRecord.Scheme, &txRecord.Points, &txRecord.RunningBalance, &txRecord.ReferenceEvent, &txRecord.CreatedAt,
			&txRecord.EventTrigger,
		); err == nil {
			list = append(list, txRecord)
		}
	}

	return list, total, nil
}

// ============================================================================
// 2. RANKS & NO AUTO-DEMOTION (FR-026, BR-017, BRULE-PRF-003)
// ============================================================================

// Mengambil seluruh master level rank diurutkan berdasarkan urutan level.
func (r *Repository) ListRanks(ctx context.Context) ([]Rank, error) {
	query := `
		SELECT id, name, min_xp, level_order, badge_icon_url, created_at
		FROM ranks
		ORDER BY level_order ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar rank: %w", err)
	}
	defer rows.Close()

	var list []Rank
	for rows.Next() {
		var rk Rank
		if err := rows.Scan(&rk.ID, &rk.Name, &rk.MinXP, &rk.LevelOrder, &rk.BadgeIconURL, &rk.CreatedAt); err == nil {
			list = append(list, rk)
		}
	}
	return list, nil
}

// Mengambil rank yang memenuhi syarat berdasarkan perolehan nilai XP kumulatif.
func (r *Repository) GetRankByXP(ctx context.Context, cumulativeXP int) (*Rank, error) {
	query := `
		SELECT id, name, min_xp, level_order, badge_icon_url, created_at
		FROM ranks
		WHERE min_xp <= $1
		ORDER BY min_xp DESC
		LIMIT 1
	`
	var rk Rank
	err := r.db.Pool.QueryRow(ctx, query, cumulativeXP).Scan(
		&rk.ID, &rk.Name, &rk.MinXP, &rk.LevelOrder, &rk.BadgeIconURL, &rk.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mengevaluasi rank berdasarkan XP: %w", err)
	}
	return &rk, nil
}

// Mengambil tingkatan rank berikutnya di atas level saat ini.
func (r *Repository) GetNextRank(ctx context.Context, currentLevelOrder int) (*Rank, error) {
	query := `
		SELECT id, name, min_xp, level_order, badge_icon_url, created_at
		FROM ranks
		WHERE level_order = $1 + 1
		LIMIT 1
	`
	var rk Rank
	err := r.db.Pool.QueryRow(ctx, query, currentLevelOrder).Scan(
		&rk.ID, &rk.Name, &rk.MinXP, &rk.LevelOrder, &rk.BadgeIconURL, &rk.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Sudah di level tertinggi
		}
		return nil, fmt.Errorf("gagal mencari rank berikutnya: %w", err)
	}
	return &rk, nil
}

// Memperbarui rank dan XP peserta magang dengan penegakan aturan permanen tanpa penurunan otomatis (No Auto-Demotion).
func (r *Repository) UpdateInternRankTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newRankID uuid.UUID, newXP int) error {
	query := `
		UPDATE interns
		SET current_rank_id = CASE
		        WHEN current_rank_id IS NULL THEN $2
		        WHEN (SELECT level_order FROM ranks WHERE id = $2) >= (SELECT level_order FROM ranks WHERE id = interns.current_rank_id) THEN $2
		        ELSE current_rank_id
		    END,
		    internship_xp = $3,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
	`
	_, err := tx.Exec(ctx, query, userID, newRankID, newXP)
	if err != nil {
		return fmt.Errorf("gagal memperbarui rank peserta: %w", err)
	}
	return nil
}

// ============================================================================
// 3. PERFORMANCE EVALUATIONS & TOP PERFORMER (FR-027, FR-028, BR-017, BR-018)
// ============================================================================

// Menyimpan dokumen evaluasi kinerja formal komposit dari supervisor ke dalam database.
func (r *Repository) CreateEvaluation(ctx context.Context, eval *PerformanceEvaluation) error {
	query := `
		INSERT INTO performance_evaluations (id, user_id, evaluator_id, batch_id, period_type, attendance_score, task_delivery_score, work_quality_score, supervisor_rubric_score, composite_performance_score, is_top_performer, notes, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP)
	`
	if eval.ID == uuid.Nil {
		eval.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, eval.ID, eval.UserID, eval.EvaluatorID, eval.BatchID, eval.PeriodType, eval.AttendanceScore, eval.TaskDeliveryScore, eval.WorkQualityScore, eval.SupervisorRubricScore, eval.CompositePerformanceScore, eval.IsTopPerformer, eval.Notes)
	if err != nil {
		return fmt.Errorf("gagal menyimpan evaluasi performa: %w", err)
	}
	return nil
}

// Mengambil dokumen evaluasi kinerja formal berdasarkan ID.
func (r *Repository) GetEvaluationByID(ctx context.Context, id uuid.UUID) (*PerformanceEvaluation, error) {
	query := `
		SELECT e.id, e.user_id, e.evaluator_id, e.batch_id, e.period_type, e.attendance_score, e.task_delivery_score, e.work_quality_score, e.supervisor_rubric_score, e.composite_performance_score, e.is_top_performer, e.notes, e.evaluated_at,
		       u.full_name, ev.full_name, b.name
		FROM performance_evaluations e
		JOIN users u ON u.id = e.user_id
		JOIN users ev ON ev.id = e.evaluator_id
		JOIN batches b ON b.id = e.batch_id
		WHERE e.id = $1
	`
	var eval PerformanceEvaluation
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&eval.ID, &eval.UserID, &eval.EvaluatorID, &eval.BatchID, &eval.PeriodType, &eval.AttendanceScore, &eval.TaskDeliveryScore, &eval.WorkQualityScore, &eval.SupervisorRubricScore, &eval.CompositePerformanceScore, &eval.IsTopPerformer, &eval.Notes, &eval.EvaluatedAt,
		&eval.UserFullName, &eval.EvaluatorFullName, &eval.BatchName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari evaluasi performa berdasarkan id: %w", err)
	}
	return &eval, nil
}

// Mengambil seluruh evaluasi kinerja pada suatu batch dan periode tertentu diurutkan berdasarkan skor komposit tertinggi.
func (r *Repository) GetEvaluationsByBatchAndPeriod(ctx context.Context, batchID uuid.UUID, periodType string) ([]PerformanceEvaluation, error) {
	query := `
		SELECT e.id, e.user_id, e.evaluator_id, e.batch_id, e.period_type, e.attendance_score, e.task_delivery_score, e.work_quality_score, e.supervisor_rubric_score, e.composite_performance_score, e.is_top_performer, e.notes, e.evaluated_at,
		       u.full_name, ev.full_name, b.name
		FROM performance_evaluations e
		JOIN users u ON u.id = e.user_id
		JOIN users ev ON ev.id = e.evaluator_id
		JOIN batches b ON b.id = e.batch_id
		WHERE e.batch_id = $1 AND e.period_type = $2
		ORDER BY e.composite_performance_score DESC, e.evaluated_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, batchID, periodType)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil evaluasi batch: %w", err)
	}
	defer rows.Close()

	var list []PerformanceEvaluation
	for rows.Next() {
		var eval PerformanceEvaluation
		if err := rows.Scan(
			&eval.ID, &eval.UserID, &eval.EvaluatorID, &eval.BatchID, &eval.PeriodType, &eval.AttendanceScore, &eval.TaskDeliveryScore, &eval.WorkQualityScore, &eval.SupervisorRubricScore, &eval.CompositePerformanceScore, &eval.IsTopPerformer, &eval.Notes, &eval.EvaluatedAt,
			&eval.UserFullName, &eval.EvaluatorFullName, &eval.BatchName,
		); err == nil {
			list = append(list, eval)
		}
	}
	return list, nil
}

// Mengosongkan status Top Performer pada batch dan periode tertentu sebelum menetapkan pemenang baru (Singular Top Performer).
func (r *Repository) ClearTopPerformerInBatchTx(ctx context.Context, tx pgx.Tx, batchID uuid.UUID, periodType string) error {
	query := `
		UPDATE performance_evaluations
		SET is_top_performer = FALSE
		WHERE batch_id = $1 AND period_type = $2 AND is_top_performer = TRUE
	`
	_, err := tx.Exec(ctx, query, batchID, periodType)
	if err != nil {
		return fmt.Errorf("gagal mereset status top performer: %w", err)
	}
	return nil
}

// Menetapkan tepat SATU peserta sebagai pemenang Top Performer dalam transaksi (BR-018).
func (r *Repository) SetTopPerformerTx(ctx context.Context, tx pgx.Tx, evalID uuid.UUID) error {
	query := `
		UPDATE performance_evaluations
		SET is_top_performer = TRUE
		WHERE id = $1
	`
	cmdTag, err := tx.Exec(ctx, query, evalID)
	if err != nil {
		return fmt.Errorf("gagal menetapkan top performer: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("evaluasi tidak ditemukan")
	}
	return nil
}

// ============================================================================
// 4. ACHIEVEMENTS & BADGES (FR-029)
// ============================================================================

// Membuat master lencana prestasi baru ke dalam katalog achievements.
func (r *Repository) CreateAchievement(ctx context.Context, a *Achievement) error {
	query := `
		INSERT INTO achievements (id, code, title, description, badge_icon_url, reward_xp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, a.ID, a.Code, a.Title, a.Description, a.BadgeIconURL, a.RewardXP)
	if err != nil {
		return fmt.Errorf("gagal membuat master achievement: %w", err)
	}
	return nil
}

// Mengambil master lencana prestasi berdasarkan kode uniknya.
func (r *Repository) GetAchievementByCode(ctx context.Context, code string) (*Achievement, error) {
	query := `
		SELECT id, code, title, description, badge_icon_url, reward_xp, created_at
		FROM achievements
		WHERE code = $1
	`
	var a Achievement
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&a.ID, &a.Code, &a.Title, &a.Description, &a.BadgeIconURL, &a.RewardXP, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari achievement berdasarkan kode: %w", err)
	}
	return &a, nil
}

// Mengambil seluruh master lencana prestasi yang terdaftar.
func (r *Repository) ListAchievements(ctx context.Context) ([]Achievement, error) {
	query := `
		SELECT id, code, title, description, badge_icon_url, reward_xp, created_at
		FROM achievements
		ORDER BY code ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar achievement: %w", err)
	}
	defer rows.Close()

	var list []Achievement
	for rows.Next() {
		var a Achievement
		if err := rows.Scan(&a.ID, &a.Code, &a.Title, &a.Description, &a.BadgeIconURL, &a.RewardXP, &a.CreatedAt); err == nil {
			list = append(list, a)
		}
	}
	return list, nil
}

// Memberikan lencana prestasi kepada pengguna secara idempoten (ON CONFLICT DO NOTHING).
func (r *Repository) UnlockAchievement(ctx context.Context, userID, achievementID uuid.UUID) (bool, error) {
	query := `
		INSERT INTO user_achievements (id, user_id, achievement_id, unlocked_at)
		VALUES (gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, achievement_id) DO NOTHING
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, userID, achievementID)
	if err != nil {
		return false, fmt.Errorf("gagal membuka achievement pengguna: %w", err)
	}
	return cmdTag.RowsAffected() > 0, nil
}

// Mengambil seluruh lencana prestasi yang telah berhasil dibuka oleh pengguna.
func (r *Repository) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]UserAchievement, error) {
	query := `
		SELECT ua.id, ua.user_id, ua.achievement_id, ua.unlocked_at,
		       a.title, a.description, a.badge_icon_url, a.reward_xp
		FROM user_achievements ua
		JOIN achievements a ON a.id = ua.achievement_id
		WHERE ua.user_id = $1
		ORDER BY ua.unlocked_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil achievement pengguna: %w", err)
	}
	defer rows.Close()

	var list []UserAchievement
	for rows.Next() {
		var ua UserAchievement
		if err := rows.Scan(&ua.ID, &ua.UserID, &ua.AchievementID, &ua.UnlockedAt, &ua.Title, &ua.Description, &ua.BadgeIconURL, &ua.RewardXP); err == nil {
			list = append(list, ua)
		}
	}
	return list, nil
}

// ============================================================================
// 5. SKILL GROWTH MATRIX (FR-030, T-061)
// ============================================================================

// Mengagregasi matriks perkembangan keahlian individu berdasarkan tugas proyek selesai dan tingkat profisiensi.
func (r *Repository) GetSkillGrowthMatrix(ctx context.Context, userID uuid.UUID) ([]SkillGrowthEntry, error) {
	query := `
		SELECT s.id, s.name, s.category, us.proficiency_level,
		       COUNT(t.id) as completed_tasks,
		       COALESCE(SUM(t.estimated_hours), 0.0) as total_hours
		FROM user_skills us
		JOIN skills s ON s.id = us.skill_id
		LEFT JOIN tasks t ON t.assignee_id = us.user_id AND t.status = 'COMPLETED'
		WHERE us.user_id = $1
		GROUP BY s.id, s.name, s.category, us.proficiency_level
		ORDER BY us.proficiency_level DESC, s.name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengagregasi matriks pertumbuhan skill: %w", err)
	}
	defer rows.Close()

	var list []SkillGrowthEntry
	for rows.Next() {
		var entry SkillGrowthEntry
		if err := rows.Scan(&entry.SkillID, &entry.SkillName, &entry.SkillCategory, &entry.ProficiencyLevel, &entry.CompletedTasks, &entry.TotalHoursWorked); err == nil {
			list = append(list, entry)
		}
	}
	return list, nil
}
