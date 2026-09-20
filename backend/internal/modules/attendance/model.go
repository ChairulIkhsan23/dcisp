package attendance

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

var WIB = time.FixedZone("WIB", 7*3600)

// 9 Jenis Event Presensi Resmi (FR-008)
const (
	EventArrived         = "ARRIVED"
	EventCheckIn         = "CHECK_IN"
	EventWorkStarted     = "WORK_STARTED"
	EventBreakStarted    = "BREAK_STARTED"
	EventBreakEnded      = "BREAK_ENDED"
	EventWorkResumed     = "WORK_RESUMED"
	EventOvertimeStarted = "OVERTIME_STARTED"
	EventOvertimeEnded   = "OVERTIME_ENDED"
	EventCheckOut        = "CHECK_OUT"
)

// 5 Metode Presensi Resmi
const (
	MethodNFC              = "NFC"
	MethodQRCode           = "QR_CODE"
	MethodAdminScanner     = "ADMIN_SCANNER"
	MethodNFCOfflineSync   = "NFC_OFFLINE_SYNC"
	MethodManualCorrection = "MANUAL_CORRECTION"
)

// Status Sesi Kerja
const (
	SessionStatusWorking = "WORKING"
	SessionStatusBreak   = "BREAK"
	SessionStatusPaused  = "PAUSED"
	SessionStatusEnded   = "ENDED"
)

// Status Permohonan Lembur
const (
	OvertimeStatusSubmitted   = "SUBMITTED"
	OvertimeStatusUnderReview = "UNDER_REVIEW"
	OvertimeStatusApproved    = "APPROVED"
	OvertimeStatusRejected    = "REJECTED"
	OvertimeStatusCompleted   = "COMPLETED"
	OvertimeStatusCancelled   = "CANCELLED"
)

// Status Permohonan Cuti
const (
	LeaveStatusSubmitted   = "SUBMITTED"
	LeaveStatusUnderReview = "UNDER_REVIEW"
	LeaveStatusApproved    = "APPROVED"
	LeaveStatusRejected    = "REJECTED"
	LeaveStatusCancelled   = "CANCELLED"
)

// Jenis Cuti
const (
	LeaveTypeSick           = "SICK"
	LeaveTypeAcademic       = "ACADEMIC"
	LeaveTypeUrgentPersonal = "URGENT_PERSONAL"
	LeaveTypeOther          = "OTHER"
)

// Status Koreksi Presensi
const (
	CorrectionStatusSubmitted   = "SUBMITTED"
	CorrectionStatusUnderReview = "UNDER_REVIEW"
	CorrectionStatusApproved    = "APPROVED"
	CorrectionStatusRejected    = "REJECTED"
)

// Tipe & Mode Terminal Pemindai
const (
	DeviceTypeESP32Terminal  = "ESP32_NFC_TERMINAL"
	DeviceTypeCameraScanner  = "CAMERA_SCANNER"
	DeviceTypeAdminTerminal  = "ADMIN_TERMINAL"
	DeviceTypeMobileOperator = "MOBILE_OPERATOR"

	TerminalModeCheckIn    = "CHECK_IN"
	TerminalModeCheckOut   = "CHECK_OUT"
	TerminalModeBreakStart = "BREAK_START"
	TerminalModeBreakEnd   = "BREAK_END"
	TerminalModeAuto       = "AUTO"
)

// 15 Master Audio Events ESP32
const (
	AudioCardRead         = "card_read"
	AudioProcessing       = "processing"
	AudioDeviceReady      = "device_ready"
	AudioCheckIn          = "check_in"
	AudioWorkStart        = "work_start"
	AudioBreakStart       = "break_start"
	AudioBreakEnd         = "break_end"
	AudioCheckOut         = "check_out"
	AudioCardUnregistered = "card_unregistered"
	AudioCardInvalid      = "card_invalid"
	AudioSaveFailed       = "save_failed"
	AudioOffline          = "offline"
	AudioOfflineSuccess   = "offline_success"
	AudioTooFrequent      = "too_frequent"
	AudioLate             = "late"
)

// Entitas Terminal Pemindai (Devices)
type Device struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	TerminalIdentifier  string     `json:"terminal_identifier" db:"terminal_identifier"`
	DeviceType          string     `json:"device_type" db:"device_type"`
	LocationName        string     `json:"location_name" db:"location_name"`
	APIKeyHash          string     `json:"-" db:"api_key_hash"`
	CurrentMode         string     `json:"current_mode" db:"current_mode"`
	FirmwareVersion     string     `json:"firmware_version" db:"firmware_version"`
	AudioCatalogVersion string     `json:"audio_catalog_version" db:"audio_catalog_version"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	LastHeartbeatAt     *time.Time `json:"last_heartbeat_at,omitempty" db:"last_heartbeat_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// Entitas Master Jadwal Kerja
type WorkSchedule struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	Name               string     `json:"name" db:"name"`
	WorkingDays        []int      `json:"working_days" db:"working_days"` // 1=Senin, 5=Jumat
	StartTime          string     `json:"start_time" db:"start_time"`     // HH:MM:SS
	EndTime            string     `json:"end_time" db:"end_time"`         // HH:MM:SS
	BreakStart         string     `json:"break_start" db:"break_start"`   // HH:MM:SS
	BreakEnd           string     `json:"break_end" db:"break_end"`       // HH:MM:SS
	GracePeriodMinutes int        `json:"grace_period_minutes" db:"grace_period_minutes"`
	OvertimePolicyID   *uuid.UUID `json:"overtime_policy_id,omitempty" db:"overtime_policy_id"`
	HolidayCalendarID  *uuid.UUID `json:"holiday_calendar_id,omitempty" db:"holiday_calendar_id"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// Entitas Hari Libur
type Holiday struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	HolidayCalendarID *uuid.UUID `json:"holiday_calendar_id,omitempty" db:"holiday_calendar_id"`
	Date              time.Time  `json:"date" db:"date"`
	Name              string     `json:"name" db:"name"`
	IsNational        bool       `json:"is_national" db:"is_national"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// Entitas Log Kehadiran Append-Only (Immutable)
type AttendanceEventLog struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	DeviceID        *uuid.UUID      `json:"device_id,omitempty" db:"device_id"`
	EventType       string          `json:"event_type" db:"event_type"`
	Timestamp       time.Time       `json:"timestamp" db:"timestamp"`
	Method          string          `json:"method" db:"method"`
	LocationContext *string         `json:"location_context,omitempty" db:"location_context"`
	SessionID       *uuid.UUID      `json:"session_id,omitempty" db:"session_id"`
	IdempotencyKey  *string         `json:"idempotency_key,omitempty" db:"idempotency_key"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`

	// Relational details
	UserFullName string `json:"user_full_name,omitempty"`
}

// Entitas Sesi Kerja Harian
type WorkSession struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	UserID                uuid.UUID  `json:"user_id" db:"user_id"`
	Date                  time.Time  `json:"date" db:"date"`
	StartTime             time.Time  `json:"start_time" db:"start_time"`
	EndTime               *time.Time `json:"end_time,omitempty" db:"end_time"`
	GrossDurationSeconds  int        `json:"gross_duration_seconds" db:"gross_duration_seconds"`
	ActiveDurationSeconds int        `json:"active_duration_seconds" db:"active_duration_seconds"`
	IdleDurationSeconds   int        `json:"idle_duration_seconds" db:"idle_duration_seconds"`
	ActiveTaskID          *uuid.UUID `json:"active_task_id,omitempty" db:"active_task_id"`
	Status                string     `json:"status" db:"status"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// Entitas Rekaman Istirahat
type Break struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	SessionID       uuid.UUID  `json:"session_id" db:"session_id"`
	StartTime       time.Time  `json:"start_time" db:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty" db:"end_time"`
	DurationSeconds int        `json:"duration_seconds" db:"duration_seconds"`
	IsAnomalyEarly  bool       `json:"is_anomaly_early" db:"is_anomaly_early"`
	IsViolationLate bool       `json:"is_violation_late" db:"is_violation_late"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// Entitas Permohonan Lembur
type OvertimeRequest struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	ProjectID           *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	TaskID              *uuid.UUID `json:"task_id,omitempty" db:"task_id"`
	Date                time.Time  `json:"date" db:"date"`
	RequestedStart      string     `json:"requested_start" db:"requested_start"`
	RequestedEnd        string     `json:"requested_end" db:"requested_end"`
	ApprovedStart       *string    `json:"approved_start,omitempty" db:"approved_start"`
	ApprovedEnd         *string    `json:"approved_end,omitempty" db:"approved_end"`
	ActualWorkedMinutes int        `json:"actual_worked_minutes" db:"actual_worked_minutes"`
	Reason              string     `json:"reason" db:"reason"`
	Status              string     `json:"status" db:"status"`
	ReviewerID          *uuid.UUID `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewNotes         *string    `json:"review_notes,omitempty" db:"review_notes"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`

	UserFullName string `json:"user_full_name,omitempty"`
}

// Entitas Permohonan Cuti
type LeaveRequest struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	LeaveType      string     `json:"leave_type" db:"leave_type"`
	StartDate      time.Time  `json:"start_date" db:"start_date"`
	EndDate        time.Time  `json:"end_date" db:"end_date"`
	Reason         string     `json:"reason" db:"reason"`
	EvidenceFileID *uuid.UUID `json:"evidence_file_id,omitempty" db:"evidence_file_id"`
	Status         string     `json:"status" db:"status"`
	ReviewerID     *uuid.UUID `json:"reviewer_id,omitempty" db:"reviewer_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	UserFullName string `json:"user_full_name,omitempty"`
}

// Entitas Koreksi Presensi Append-Only
type AttendanceCorrection struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	AttendanceEventID *uuid.UUID `json:"attendance_event_id,omitempty" db:"attendance_event_id"`
	TargetDate        time.Time  `json:"target_date" db:"target_date"`
	ProposedEventType string     `json:"proposed_event_type" db:"proposed_event_type"`
	ProposedTimestamp time.Time  `json:"proposed_timestamp" db:"proposed_timestamp"`
	Reason            string     `json:"reason" db:"reason"`
	EvidenceFileID    *uuid.UUID `json:"evidence_file_id,omitempty" db:"evidence_file_id"`
	Status            string     `json:"status" db:"status"`
	ReviewerID        *uuid.UUID `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`

	UserFullName string `json:"user_full_name,omitempty"`
}

// ============================================================================
// DTOs (REQUEST / RESPONSE)
// ============================================================================

type TerminalTapRequest struct {
	CardUID        string `json:"card_uid"`
	QRToken        string `json:"qr_token"`
	Timestamp      int64  `json:"timestamp"`     // Unix timestamp RTC
	TerminalMode   string `json:"terminal_mode"` // AUTO, CHECK_IN, CHECK_OUT, BREAK_START, BREAK_END
	IdempotencyKey string `json:"idempotency_key"`
}

type TerminalTapResponse struct {
	Status      string                 `json:"status"` // SUCCESS, DEBOUNCED, ERROR
	EventType   string                 `json:"event_type"`
	IsLate      bool                   `json:"is_late"`
	LateTier    int                    `json:"late_tier,omitempty"` // 1, 2, 3
	LateMinutes int                    `json:"late_minutes,omitempty"`
	AudioEvent  string                 `json:"audio_event"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

type OfflineSyncRecord struct {
	TerminalIdentifier string `json:"terminal_identifier" binding:"required"`
	CardUID            string `json:"card_uid" binding:"required"`
	Timestamp          int64  `json:"timestamp" binding:"required"`
	EventType          string `json:"event_type"`
	IdempotencyKey     string `json:"idempotency_key" binding:"required"`
}

type OfflineSyncRequest struct {
	Records []OfflineSyncRecord `json:"records" binding:"required"`
}

type OfflineSyncResponse struct {
	SyncedCount  int      `json:"synced_count"`
	SkippedCount int      `json:"skipped_count"`
	Errors       []string `json:"errors,omitempty"`
}

type GenerateQRResponse struct {
	QRToken          string `json:"qr_token"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
	ExpiresAt        string `json:"expires_at"`
}

type AudioCatalogItem struct {
	Key         string `json:"key"`
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	SHA256Hash  string `json:"sha256_hash"`
	Description string `json:"description"`
}

type AudioCatalogResponse struct {
	Version string             `json:"version"`
	BaseURL string             `json:"base_url"`
	Items   []AudioCatalogItem `json:"items"`
}

type StartWorkSessionRequest struct {
	TaskID *uuid.UUID `json:"task_id"`
}

type BreakActionRequest struct {
	Action string `json:"action" binding:"required"` // START atau RESUME
}

type EndWorkSessionRequest struct {
	WorkSummary *string `json:"work_summary"`
}

type CreateOvertimeRequest struct {
	ProjectID      *uuid.UUID `json:"project_id"`
	TaskID         *uuid.UUID `json:"task_id"`
	Date           string     `json:"date" binding:"required"`            // YYYY-MM-DD
	RequestedStart string     `json:"requested_start" binding:"required"` // HH:MM
	RequestedEnd   string     `json:"requested_end" binding:"required"`   // HH:MM
	Reason         string     `json:"reason" binding:"required"`
}

type ReviewOvertimeRequest struct {
	Action        string  `json:"action" binding:"required"` // APPROVE atau REJECT
	ApprovedStart *string `json:"approved_start"`
	ApprovedEnd   *string `json:"approved_end"`
	ReviewNotes   *string `json:"review_notes"`
}

type CreateLeaveRequest struct {
	LeaveType      string     `json:"leave_type" binding:"required"`
	StartDate      string     `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate        string     `json:"end_date" binding:"required"`   // YYYY-MM-DD
	Reason         string     `json:"reason" binding:"required"`
	EvidenceFileID *uuid.UUID `json:"evidence_file_id"`
}

type ReviewLeaveRequest struct {
	Action string `json:"action" binding:"required"` // APPROVE atau REJECT
}

type CreateCorrectionRequest struct {
	TargetDate        string     `json:"target_date" binding:"required"` // YYYY-MM-DD
	ProposedEventType string     `json:"proposed_event_type" binding:"required"`
	ProposedTimestamp string     `json:"proposed_timestamp" binding:"required"` // ISO8601
	Reason            string     `json:"reason" binding:"required,min=20"`
	EvidenceFileID    *uuid.UUID `json:"evidence_file_id"`
}

type ReviewCorrectionRequest struct {
	Action string `json:"action" binding:"required"` // APPROVE atau REJECT
}

type RegisterDeviceRequest struct {
	TerminalIdentifier string `json:"terminal_identifier" binding:"required"`
	DeviceType         string `json:"device_type" binding:"required"`
	LocationName       string `json:"location_name" binding:"required"`
	APIKey             string `json:"api_key" binding:"required,min=16"`
	CurrentMode        string `json:"current_mode"`
}

type DeviceHeartbeatRequest struct {
	FirmwareVersion     string `json:"firmware_version"`
	AudioCatalogVersion string `json:"audio_catalog_version"`
	CurrentMode         string `json:"current_mode"`
}
