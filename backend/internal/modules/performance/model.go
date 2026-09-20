package performance

import (
	"time"

	"github.com/google/uuid"
)

// 3 Skema Jalur XP Terisolasi (FR-025, BR-004)
const (
	XPSchemeInternship = "INTERNSHIP_XP"
	XPSchemeProject    = "PROJECT_XP"
	XPSchemeAlumni     = "ALUMNI_CONTRIBUTION"
)

// Periode Evaluasi Performa Formal (FR-027)
const (
	PeriodWeekly     = "WEEKLY"
	PeriodMonthly    = "MONTHLY"
	PeriodEndOfBatch = "END_OF_BATCH"
	PeriodCustom     = "CUSTOM"
)

// Entitas Master Aturan XP (xp_rules)
type XPRule struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	PolicyID     *uuid.UUID `json:"policy_id,omitempty" db:"policy_id"`
	EventTrigger string     `json:"event_trigger" db:"event_trigger"`
	XPValue      int        `json:"xp_value" db:"xp_value"` // Positif atau negatif
	XPScheme     string     `json:"xp_scheme" db:"xp_scheme"`
	Description  *string    `json:"description,omitempty" db:"description"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// Entitas Buku Besar Transaksi Mutasi XP Append-Only (Immutable - xp_transactions)
type XPTransaction struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	XPRuleID       *uuid.UUID `json:"xp_rule_id,omitempty" db:"xp_rule_id"`
	Scheme         string     `json:"scheme" db:"scheme"`
	Points         int        `json:"points" db:"points"`
	RunningBalance int        `json:"running_balance" db:"running_balance"`
	ReferenceEvent *string    `json:"reference_event,omitempty" db:"reference_event"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`

	EventTrigger *string `json:"event_trigger,omitempty"`
}

// Entitas Tingkat Rank Gamifikasi (ranks)
type Rank struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	MinXP        int       `json:"min_xp" db:"min_xp"`
	LevelOrder   int       `json:"level_order" db:"level_order"`
	BadgeIconURL *string   `json:"badge_icon_url,omitempty" db:"badge_icon_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Entitas Lencana Prestasi (achievements)
type Achievement struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Code         string    `json:"code" db:"code"`
	Title        string    `json:"title" db:"title"`
	Description  string    `json:"description" db:"description"`
	BadgeIconURL *string   `json:"badge_icon_url,omitempty" db:"badge_icon_url"`
	RewardXP     int       `json:"reward_xp" db:"reward_xp"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Entitas Lencana Prestasi Pengguna (user_achievements)
type UserAchievement struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	AchievementID uuid.UUID `json:"achievement_id" db:"achievement_id"`
	UnlockedAt    time.Time `json:"unlocked_at" db:"unlocked_at"`

	Title        string  `json:"title,omitempty"`
	Description  string  `json:"description,omitempty"`
	BadgeIconURL *string `json:"badge_icon_url,omitempty"`
	RewardXP     int     `json:"reward_xp,omitempty"`
}

// Entitas Evaluasi Kinerja Formal Supervisor (FR-027, BR-017)
type PerformanceEvaluation struct {
	ID                        uuid.UUID `json:"id" db:"id"`
	UserID                    uuid.UUID `json:"user_id" db:"user_id"`
	EvaluatorID               uuid.UUID `json:"evaluator_id" db:"evaluator_id"`
	BatchID                   uuid.UUID `json:"batch_id" db:"batch_id"`
	PeriodType                string    `json:"period_type" db:"period_type"`
	AttendanceScore           float64   `json:"attendance_score" db:"attendance_score"`
	TaskDeliveryScore         float64   `json:"task_delivery_score" db:"task_delivery_score"`
	WorkQualityScore          float64   `json:"work_quality_score" db:"work_quality_score"`
	SupervisorRubricScore     float64   `json:"supervisor_rubric_score" db:"supervisor_rubric_score"`
	CompositePerformanceScore float64   `json:"composite_performance_score" db:"composite_performance_score"`
	IsTopPerformer            bool      `json:"is_top_performer" db:"is_top_performer"`
	Notes                     *string   `json:"notes,omitempty" db:"notes"`
	EvaluatedAt               time.Time `json:"evaluated_at" db:"evaluated_at"`

	UserFullName      string `json:"user_full_name,omitempty"`
	EvaluatorFullName string `json:"evaluator_full_name,omitempty"`
	BatchName         string `json:"batch_name,omitempty"`
}

// ============================================================================
// DTOs (REQUEST / RESPONSE)
// ============================================================================

type RecordXPMutationRequest struct {
	UserID         uuid.UUID `json:"user_id" binding:"required"`
	EventTrigger   string    `json:"event_trigger" binding:"required"`
	PointsOverride *int      `json:"points_override"` // Opsional jika ingin override aturan
	ReferenceEvent string    `json:"reference_event" binding:"required"`
}

type XPBalanceResponse struct {
	UserID               uuid.UUID `json:"user_id"`
	InternshipXP         int       `json:"internship_xp"`
	ProjectXP            int       `json:"project_xp"`
	AlumniContributionXP int       `json:"alumni_contribution_xp"`
	CurrentRank          *Rank     `json:"current_rank,omitempty"`
	NextRank             *Rank     `json:"next_rank,omitempty"`
	XPToNextRank         int       `json:"xp_to_next_rank"`
}

type CreateEvaluationRequest struct {
	UserID                uuid.UUID `json:"user_id" binding:"required"`
	BatchID               uuid.UUID `json:"batch_id" binding:"required"`
	PeriodType            string    `json:"period_type" binding:"required"`
	AttendanceScore       float64   `json:"attendance_score" binding:"required,min=0,max=100"`
	TaskDeliveryScore     float64   `json:"task_delivery_score" binding:"required,min=0,max=100"`
	WorkQualityScore      float64   `json:"work_quality_score" binding:"required,min=0,max=100"`
	SupervisorRubricScore float64   `json:"supervisor_rubric_score" binding:"required,min=0,max=100"`
	Notes                 *string   `json:"notes"`
}

type TopPerformerResponse struct {
	BatchID                   uuid.UUID `json:"batch_id"`
	WinnerUserID              uuid.UUID `json:"winner_user_id"`
	WinnerFullName            string    `json:"winner_full_name"`
	CompositePerformanceScore float64   `json:"composite_performance_score"`
	PeriodType                string    `json:"period_type"`
	EvaluationDate            string    `json:"evaluation_date"`
}

type CreateAchievementRequest struct {
	Code         string  `json:"code" binding:"required"`
	Title        string  `json:"title" binding:"required"`
	Description  string  `json:"description" binding:"required"`
	BadgeIconURL *string `json:"badge_icon_url"`
	RewardXP     int     `json:"reward_xp" binding:"min=0"`
}

type SkillGrowthEntry struct {
	SkillID          uuid.UUID `json:"skill_id"`
	SkillName        string    `json:"skill_name"`
	SkillCategory    string    `json:"skill_category"`
	ProficiencyLevel int       `json:"proficiency_level"`
	CompletedTasks   int       `json:"completed_tasks"`
	TotalHoursWorked float64   `json:"total_hours_worked"`
}

type SkillGrowthMatrixResponse struct {
	UserID uuid.UUID          `json:"user_id"`
	Skills []SkillGrowthEntry `json:"skills"`
}
