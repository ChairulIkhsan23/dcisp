package projects

import (
	"time"

	"github.com/google/uuid"
)

var WIB = time.FixedZone("WIB", 7*3600)

// Visibilitas Proyek (FR-016)
const (
	VisibilityInternOnly = "INTERN_ONLY"
	VisibilityPublic     = "PUBLIC"
	VisibilityPrivate    = "PRIVATE"
)

// Status Siklus Hidup Proyek
const (
	ProjectStatusDraft      = "DRAFT"
	ProjectStatusPublished  = "PUBLISHED"
	ProjectStatusInProgress = "IN_PROGRESS"
	ProjectStatusCompleted  = "COMPLETED"
	ProjectStatusCancelled  = "CANCELLED"
)

// Status Lamaran Proyek (FR-017)
const (
	AppStatusApplied     = "APPLIED"
	AppStatusUnderReview = "UNDER_REVIEW"
	AppStatusShortlisted = "SHORTLISTED"
	AppStatusAccepted    = "ACCEPTED"
	AppStatusRejected    = "REJECTED"
	AppStatusWithdrawn   = "WITHDRAWN"
)

// Peran Anggota Tim Proyek (FR-019)
const (
	ProjectRoleOwner      = "OWNER"
	ProjectRoleManager    = "MANAGER"
	ProjectRoleSupervisor = "SUPERVISOR"
	ProjectRoleMember     = "MEMBER"
)

// Status Milestone Proyek (FR-020)
const (
	MilestoneStatusPending    = "PENDING"
	MilestoneStatusInProgress = "IN_PROGRESS"
	MilestoneStatusCompleted  = "COMPLETED"
)

// Status Tugas Kanban (FR-021)
const (
	TaskStatusTodo       = "TODO"
	TaskStatusInProgress = "IN_PROGRESS"
	TaskStatusInReview   = "IN_REVIEW"
	TaskStatusCompleted  = "COMPLETED"
	TaskStatusBlocked    = "BLOCKED"
	TaskStatusCancelled  = "CANCELLED"
)

// Prioritas Tugas Kanban
const (
	TaskPriorityLow      = "LOW"
	TaskPriorityMedium   = "MEDIUM"
	TaskPriorityHigh     = "HIGH"
	TaskPriorityCritical = "CRITICAL"
)

// Status Laporan Pekerjaan (FR-022)
const (
	ReportStatusSubmitted        = "SUBMITTED"
	ReportStatusUnderReview      = "UNDER_REVIEW"
	ReportStatusApproved         = "APPROVED"
	ReportStatusRevisionRequired = "REVISION_REQUIRED"
)

// Tipe Artefak Bukti Kerja (FR-023)
const (
	EvidenceTypeGitCommit  = "GIT_COMMIT"
	EvidenceTypeScreenshot = "SCREENSHOT"
	EvidenceTypeURL        = "URL"
	EvidenceTypeDocument   = "DOCUMENT"
	EvidenceTypeAttachment = "ATTACHMENT"
)

// Entitas Master Proyek (DATA-003)
type Project struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	Title          string      `json:"title" db:"title"`
	Description    string      `json:"description" db:"description"`
	OwnerID        uuid.UUID   `json:"owner_id" db:"owner_id"`
	Visibility     string      `json:"visibility" db:"visibility"`
	RequiredSkills []uuid.UUID `json:"required_skills" db:"required_skills"`
	Capacity       int         `json:"capacity" db:"capacity"`
	AcceptedCount  int         `json:"accepted_count" db:"accepted_count"`
	Deadline       time.Time   `json:"deadline" db:"deadline"`
	BountyPool     float64     `json:"bounty_pool" db:"bounty_pool"`
	Status         string      `json:"status" db:"status"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`

	OwnerFullName string `json:"owner_full_name,omitempty"`
}

// Entitas Berkas Lamaran Proyek
type ProjectApplication struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	ProjectID            uuid.UUID `json:"project_id" db:"project_id"`
	UserID               uuid.UUID `json:"user_id" db:"user_id"`
	CoverLetter          *string   `json:"cover_letter,omitempty" db:"cover_letter"`
	SkillMatchPercentage *float64  `json:"skill_match_percentage,omitempty" db:"skill_match_percentage"`
	Status               string    `json:"status" db:"status"`
	AppliedAt            time.Time `json:"applied_at" db:"applied_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`

	ApplicantFullName string `json:"applicant_full_name,omitempty"`
	ApplicantEmail    string `json:"applicant_email,omitempty"`
}

// Entitas Anggota Tim Proyek (Three-Layer Contribution)
type ProjectTeamMember struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	ProjectID              uuid.UUID `json:"project_id" db:"project_id"`
	UserID                 uuid.UUID `json:"user_id" db:"user_id"`
	ProjectRole            string    `json:"project_role" db:"project_role"`
	Responsibility         *string   `json:"responsibility,omitempty" db:"responsibility"`
	PlannedContributionPct float64   `json:"planned_contribution_pct" db:"planned_contribution_pct"`
	ActualContributionPct  *float64  `json:"actual_contribution_pct,omitempty" db:"actual_contribution_pct"`
	FinalContributionPct   *float64  `json:"final_contribution_pct,omitempty" db:"final_contribution_pct"`
	IsLocked               bool      `json:"is_locked" db:"is_locked"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`

	MemberFullName string `json:"member_full_name,omitempty"`
	MemberEmail    string `json:"member_email,omitempty"`
}

// Entitas Milestone Target Fase Proyek
type Milestone struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ProjectID   uuid.UUID `json:"project_id" db:"project_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description,omitempty" db:"description"`
	Deadline    time.Time `json:"deadline" db:"deadline"`
	WeightPct   float64   `json:"weight_pct" db:"weight_pct"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Entitas Tugas Kanban Proyek
type Task struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	ProjectID        uuid.UUID  `json:"project_id" db:"project_id"`
	MilestoneID      *uuid.UUID `json:"milestone_id,omitempty" db:"milestone_id"`
	AssigneeID       *uuid.UUID `json:"assignee_id,omitempty" db:"assignee_id"`
	Title            string     `json:"title" db:"title"`
	Description      *string    `json:"description,omitempty" db:"description"`
	EstimatedHours   float64    `json:"estimated_hours" db:"estimated_hours"`
	DifficultyWeight int        `json:"difficulty_weight" db:"difficulty_weight"`
	Priority         string     `json:"priority" db:"priority"`
	Deadline         time.Time  `json:"deadline" db:"deadline"`
	Status           string     `json:"status" db:"status"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	AssigneeFullName *string `json:"assignee_full_name,omitempty"`
	MilestoneTitle   *string `json:"milestone_title,omitempty"`
}

// Entitas Laporan Kemajuan Pekerjaan (DATA-004)
type WorkReport struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	TaskID             uuid.UUID `json:"task_id" db:"task_id"`
	UserID             uuid.UUID `json:"user_id" db:"user_id"`
	Date               time.Time `json:"date" db:"date"`
	ProgressPercentage int       `json:"progress_percentage" db:"progress_percentage"`
	WhatIDid           string    `json:"what_i_did" db:"what_i_did"`
	EvidenceType       string    `json:"evidence_type" db:"evidence_type"`
	EvidenceURLOrKey   string    `json:"evidence_url_or_key" db:"evidence_url_or_key"`
	Problems           *string   `json:"problems,omitempty" db:"problems"`
	NextActions        *string   `json:"next_actions,omitempty" db:"next_actions"`
	Status             string    `json:"status" db:"status"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`

	TaskTitle      string `json:"task_title,omitempty"`
	AuthorFullName string `json:"author_full_name,omitempty"`
}

// Entitas Artefak Bukti Deliverable
type Evidence struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	SubmissionID   uuid.UUID  `json:"submission_id" db:"submission_id"`
	FileMetadataID *uuid.UUID `json:"file_metadata_id,omitempty" db:"file_metadata_id"`
	EvidenceType   string     `json:"evidence_type" db:"evidence_type"`
	ExternalURL    *string    `json:"external_url,omitempty" db:"external_url"`
	Description    *string    `json:"description,omitempty" db:"description"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// ============================================================================
// DTOs (REQUEST / RESPONSE)
// ============================================================================

type CreateProjectRequest struct {
	Title          string      `json:"title" binding:"required"`
	Description    string      `json:"description" binding:"required"`
	Visibility     string      `json:"visibility" binding:"required"` // INTERN_ONLY, PUBLIC, PRIVATE
	RequiredSkills []uuid.UUID `json:"required_skills"`
	Capacity       int         `json:"capacity" binding:"required,min=1"`
	Deadline       string      `json:"deadline" binding:"required"` // ISO8601 / RFC3339
	BountyPool     float64     `json:"bounty_pool" binding:"min=0"`
	Status         string      `json:"status"` // DRAFT, PUBLISHED
}

type UpdateProjectRequest struct {
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	Visibility     string      `json:"visibility"`
	RequiredSkills []uuid.UUID `json:"required_skills"`
	Capacity       *int        `json:"capacity"`
	Deadline       string      `json:"deadline"`
	BountyPool     *float64    `json:"bounty_pool"`
	Status         string      `json:"status"`
}

type ApplyProjectRequest struct {
	CoverLetter string `json:"cover_letter"`
}

type ReviewApplicationRequest struct {
	Action string `json:"action" binding:"required"` // ACCEPT atau REJECT
}

type AddTeamMemberRequest struct {
	UserID                 uuid.UUID `json:"user_id" binding:"required"`
	ProjectRole            string    `json:"project_role" binding:"required"`
	Responsibility         *string   `json:"responsibility"`
	PlannedContributionPct float64   `json:"planned_contribution_pct" binding:"min=0,max=100"`
}

type MemberPlannedContribution struct {
	UserID                 uuid.UUID `json:"user_id" binding:"required"`
	PlannedContributionPct float64   `json:"planned_contribution_pct" binding:"required,min=0,max=100"`
}

type FinalizePlannedContributionRequest struct {
	Contributions []MemberPlannedContribution `json:"contributions" binding:"required"`
}

type CreateMilestoneRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Deadline    string  `json:"deadline" binding:"required"` // ISO8601
	WeightPct   float64 `json:"weight_pct" binding:"required,min=0,max=100"`
}

type UpdateMilestoneRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	Deadline    string   `json:"deadline"`
	WeightPct   *float64 `json:"weight_pct"`
	Status      string   `json:"status"`
}

type CreateTaskRequest struct {
	MilestoneID      *uuid.UUID `json:"milestone_id"`
	AssigneeID       *uuid.UUID `json:"assignee_id"`
	Title            string     `json:"title" binding:"required"`
	Description      *string    `json:"description"`
	EstimatedHours   float64    `json:"estimated_hours" binding:"required,min=0"`
	DifficultyWeight int        `json:"difficulty_weight" binding:"required,min=1,max=5"`
	Priority         string     `json:"priority" binding:"required"`
	Deadline         string     `json:"deadline" binding:"required"` // ISO8601
}

type UpdateTaskRequest struct {
	MilestoneID      *uuid.UUID `json:"milestone_id"`
	AssigneeID       *uuid.UUID `json:"assignee_id"`
	Title            string     `json:"title"`
	Description      *string    `json:"description"`
	EstimatedHours   *float64   `json:"estimated_hours"`
	DifficultyWeight *int       `json:"difficulty_weight"`
	Priority         string     `json:"priority"`
	Deadline         string     `json:"deadline"`
}

type ChangeTaskStatusRequest struct {
	Status string `json:"status" binding:"required"` // State machine target
}

type SubmitWorkReportRequest struct {
	ProgressPercentage int     `json:"progress_percentage" binding:"required,min=0,max=100"`
	WhatIDid           string  `json:"what_i_did" binding:"required,min=20"`
	EvidenceType       string  `json:"evidence_type" binding:"required"`
	EvidenceURLOrKey   string  `json:"evidence_url_or_key" binding:"required"`
	Problems           *string `json:"problems"`
	NextActions        *string `json:"next_actions"`
}

type ReviewWorkReportRequest struct {
	Action string `json:"action" binding:"required"` // APPROVE atau REVISION_REQUIRED
	Notes  string `json:"notes"`
}

type AddEvidenceRequest struct {
	FileMetadataID *uuid.UUID `json:"file_metadata_id"`
	EvidenceType   string     `json:"evidence_type" binding:"required"`
	ExternalURL    *string    `json:"external_url"`
	Description    *string    `json:"description"`
}

type MemberFinalContribution struct {
	UserID               uuid.UUID `json:"user_id" binding:"required"`
	FinalContributionPct float64   `json:"final_contribution_pct" binding:"required,min=0,max=100"`
}

type FinalizeContributionRequest struct {
	FinalContributions []MemberFinalContribution `json:"final_contributions" binding:"required"`
	EvaluationNotes    string                    `json:"evaluation_notes" binding:"required"`
}

type SkillMatchResponse struct {
	ProjectID      uuid.UUID   `json:"project_id"`
	UserID         uuid.UUID   `json:"user_id"`
	MatchScore     float64     `json:"match_score"` // 0 - 100%
	MatchedSkills  []uuid.UUID `json:"matched_skills"`
	MissingSkills  []uuid.UUID `json:"missing_skills"`
	Recommendation string      `json:"recommendation"`
	IsAdvisoryOnly bool        `json:"is_advisory_only"`
}
