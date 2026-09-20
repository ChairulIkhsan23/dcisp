package people

import (
	"time"

	"github.com/google/uuid"
)

// Status siklus hidup peserta magang (7 status resmi sesuai FR-002)
const (
	InternStatusApplicant  = "APPLICANT"
	InternStatusOnboarding = "ONBOARDING"
	InternStatusActive     = "ACTIVE"
	InternStatusOnLeave    = "ON_LEAVE"
	InternStatusSuspended  = "SUSPENDED"
	InternStatusGraduated  = "GRADUATED"
	InternStatusTerminated = "TERMINATED"
)

// Status kelompok angkatan (Batch)
const (
	BatchStatusDraft     = "DRAFT"
	BatchStatusActive    = "ACTIVE"
	BatchStatusCompleted = "COMPLETED"
	BatchStatusArchived  = "ARCHIVED"
)

type Institution struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	Address       *string    `json:"address,omitempty" db:"address"`
	ContactPerson *string    `json:"contact_person,omitempty" db:"contact_person"`
	Email         *string    `json:"email,omitempty" db:"email"`
	Phone         *string    `json:"phone,omitempty" db:"phone"`
	ExternalID    *string    `json:"external_id,omitempty" db:"external_id"`
	Source        *string    `json:"source,omitempty" db:"source"`
	SyncedAt      *time.Time `json:"synced_at,omitempty" db:"synced_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type Batch struct {
	ID        uuid.UUID `json:"id" db:"id"`
	BatchCode string    `json:"batch_code" db:"batch_code"`
	Name      string    `json:"name" db:"name"`
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`
	Quota     int       `json:"quota" db:"quota"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type BatchFund struct {
	ID               uuid.UUID `json:"id" db:"id"`
	BatchID          uuid.UUID `json:"batch_id" db:"batch_id"`
	TotalAccumulated float64   `json:"total_accumulated" db:"total_accumulated"`
	CurrentBalance   float64   `json:"current_balance" db:"current_balance"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type Intern struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	BatchID       uuid.UUID  `json:"batch_id" db:"batch_id"`
	InstitutionID *uuid.UUID `json:"institution_id,omitempty" db:"institution_id"`
	IDNumber      *string    `json:"id_number,omitempty" db:"id_number"`
	MentorID      *uuid.UUID `json:"mentor_id,omitempty" db:"mentor_id"`
	Status        string     `json:"status" db:"status"`
	JoinDate      time.Time  `json:"join_date" db:"join_date"`
	EndDate       time.Time  `json:"end_date" db:"end_date"`
	CurrentRankID *uuid.UUID `json:"current_rank_id,omitempty" db:"current_rank_id"`
	InternshipXP  int        `json:"internship_xp" db:"internship_xp"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`

	// Relational details for response
	UserFullName    string  `json:"user_full_name,omitempty"`
	UserEmail       string  `json:"user_email,omitempty"`
	BatchName       string  `json:"batch_name,omitempty"`
	InstitutionName *string `json:"institution_name,omitempty"`
	MentorFullName  *string `json:"mentor_full_name,omitempty"`
	RankName        *string `json:"rank_name,omitempty"`
}

type Alumni struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	BatchID         uuid.UUID  `json:"batch_id" db:"batch_id"`
	GraduationDate  time.Time  `json:"graduation_date" db:"graduation_date"`
	AlumniXP        int        `json:"alumni_xp" db:"alumni_xp"`
	CertificateID   *uuid.UUID `json:"certificate_id,omitempty" db:"certificate_id"`
	IsPublicProfile bool       `json:"is_public_profile" db:"is_public_profile"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	// Relational details for response
	UserFullName string `json:"user_full_name,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
	BatchName    string `json:"batch_name,omitempty"`
}

type Skill struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Category    string    `json:"category" db:"category"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type UserSkill struct {
	ID               uuid.UUID `json:"id" db:"id"`
	UserID           uuid.UUID `json:"user_id" db:"user_id"`
	SkillID          uuid.UUID `json:"skill_id" db:"skill_id"`
	ProficiencyLevel int       `json:"proficiency_level" db:"proficiency_level"` // 1 s/d 5
	CreatedAt        time.Time `json:"created_at" db:"created_at"`

	SkillName     string  `json:"skill_name,omitempty"`
	SkillCategory string  `json:"skill_category,omitempty"`
	Description   *string `json:"description,omitempty"`
}

// DTOs for requests
type CreateInstitutionRequest struct {
	Name          string  `json:"name" binding:"required"`
	Address       *string `json:"address"`
	ContactPerson *string `json:"contact_person"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
}

type UpdateInstitutionRequest struct {
	Name          string  `json:"name"`
	Address       *string `json:"address"`
	ContactPerson *string `json:"contact_person"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
}

type CreateBatchRequest struct {
	BatchCode string `json:"batch_code" binding:"required"`
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate   string `json:"end_date" binding:"required"`   // YYYY-MM-DD
	Quota     int    `json:"quota" binding:"required,min=1"`
	Status    string `json:"status"` // DRAFT, ACTIVE
}

type UpdateBatchRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Quota     *int   `json:"quota"`
	Status    string `json:"status"`
}

type RegisterInternRequest struct {
	UserID        uuid.UUID  `json:"user_id" binding:"required"`
	BatchID       uuid.UUID  `json:"batch_id" binding:"required"`
	InstitutionID *uuid.UUID `json:"institution_id"`
	IDNumber      *string    `json:"id_number"`
	JoinDate      string     `json:"join_date" binding:"required"` // YYYY-MM-DD
	EndDate       string     `json:"end_date" binding:"required"`  // YYYY-MM-DD
}

type ChangeInternStatusRequest struct {
	Status string  `json:"status" binding:"required"`
	Reason *string `json:"reason"`
}

type AssignMentorRequest struct {
	MentorID uuid.UUID `json:"mentor_id" binding:"required"`
}

type AssignBatchRequest struct {
	BatchID uuid.UUID `json:"batch_id" binding:"required"`
}

type CreateSkillRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category" binding:"required"`
	Description *string `json:"description"`
}

type AssignUserSkillRequest struct {
	SkillID          uuid.UUID `json:"skill_id" binding:"required"`
	ProficiencyLevel int       `json:"proficiency_level" binding:"required,min=1,max=5"`
}

type UpdateUserSkillRequest struct {
	ProficiencyLevel int `json:"proficiency_level" binding:"required,min=1,max=5"`
}
