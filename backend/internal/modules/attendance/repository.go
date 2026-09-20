package attendance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.PostgresDB
}

// Menginisialisasi instance baru repository attendance and workforce management.
func NewRepository(db *database.PostgresDB) *Repository {
	return &Repository{db: db}
}

// ============================================================================
// 1. DEVICES & TERMINALS
// ============================================================================

// Mendaftarkan perangkat pemindai terminal baru ke dalam tabel devices.
func (r *Repository) CreateDevice(ctx context.Context, dev *Device) error {
	query := `
		INSERT INTO devices (id, terminal_identifier, device_type, location_name, api_key_hash, current_mode, firmware_version, audio_catalog_version, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if dev.ID == uuid.Nil {
		dev.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, dev.ID, dev.TerminalIdentifier, dev.DeviceType, dev.LocationName, dev.APIKeyHash, dev.CurrentMode, dev.FirmwareVersion, dev.AudioCatalogVersion, dev.IsActive)
	if err != nil {
		return fmt.Errorf("gagal mendaftarkan perangkat terminal: %w", err)
	}
	return nil
}

// Mengambil data terminal pemindai berdasarkan pengenal unik string.
func (r *Repository) GetDeviceByIdentifier(ctx context.Context, identifier string) (*Device, error) {
	query := `
		SELECT id, terminal_identifier, device_type, location_name, api_key_hash, current_mode, firmware_version, audio_catalog_version, is_active, last_heartbeat_at, created_at, updated_at
		FROM devices
		WHERE terminal_identifier = $1
	`
	var dev Device
	err := r.db.Pool.QueryRow(ctx, query, identifier).Scan(
		&dev.ID, &dev.TerminalIdentifier, &dev.DeviceType, &dev.LocationName, &dev.APIKeyHash, &dev.CurrentMode, &dev.FirmwareVersion, &dev.AudioCatalogVersion, &dev.IsActive, &dev.LastHeartbeatAt, &dev.CreatedAt, &dev.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari terminal berdasarkan pengenal: %w", err)
	}
	return &dev, nil
}

// Mengambil data terminal pemindai berdasarkan ID UUID.
func (r *Repository) GetDeviceByID(ctx context.Context, id uuid.UUID) (*Device, error) {
	query := `
		SELECT id, terminal_identifier, device_type, location_name, api_key_hash, current_mode, firmware_version, audio_catalog_version, is_active, last_heartbeat_at, created_at, updated_at
		FROM devices
		WHERE id = $1
	`
	var dev Device
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&dev.ID, &dev.TerminalIdentifier, &dev.DeviceType, &dev.LocationName, &dev.APIKeyHash, &dev.CurrentMode, &dev.FirmwareVersion, &dev.AudioCatalogVersion, &dev.IsActive, &dev.LastHeartbeatAt, &dev.CreatedAt, &dev.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari terminal berdasarkan id: %w", err)
	}
	return &dev, nil
}

// Memperbarui telemetri detak jantung (heartbeat) terminal pemindai.
func (r *Repository) UpdateDeviceHeartbeat(ctx context.Context, id uuid.UUID, firmware, audio, mode string) error {
	query := `
		UPDATE devices
		SET firmware_version = COALESCE(NULLIF($2, ''), firmware_version),
		    audio_catalog_version = COALESCE(NULLIF($3, ''), audio_catalog_version),
		    current_mode = COALESCE(NULLIF($4, ''), current_mode),
		    last_heartbeat_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id, firmware, audio, mode)
	if err != nil {
		return fmt.Errorf("gagal memperbarui heartbeat terminal: %w", err)
	}
	return nil
}

// Mengambil daftar seluruh perangkat terminal yang terdaftar di sistem.
func (r *Repository) ListDevices(ctx context.Context) ([]Device, error) {
	query := `
		SELECT id, terminal_identifier, device_type, location_name, current_mode, firmware_version, audio_catalog_version, is_active, last_heartbeat_at, created_at, updated_at
		FROM devices
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar terminal: %w", err)
	}
	defer rows.Close()

	var list []Device
	for rows.Next() {
		var dev Device
		if err := rows.Scan(&dev.ID, &dev.TerminalIdentifier, &dev.DeviceType, &dev.LocationName, &dev.CurrentMode, &dev.FirmwareVersion, &dev.AudioCatalogVersion, &dev.IsActive, &dev.LastHeartbeatAt, &dev.CreatedAt, &dev.UpdatedAt); err == nil {
			list = append(list, dev)
		}
	}
	return list, nil
}

// ============================================================================
// 2. SCHEDULES & HOLIDAYS
// ============================================================================

// Mengambil konfigurasi jadwal kerja acuan aktif pertama dari database.
func (r *Repository) GetActiveSchedule(ctx context.Context) (*WorkSchedule, error) {
	query := `
		SELECT id, name, working_days, start_time::text, end_time::text, break_start::text, break_end::text, grace_period_minutes, overtime_policy_id, holiday_calendar_id, created_at, updated_at
		FROM work_schedules
		ORDER BY created_at ASC
		LIMIT 1
	`
	var ws WorkSchedule
	err := r.db.Pool.QueryRow(ctx, query).Scan(
		&ws.ID, &ws.Name, &ws.WorkingDays, &ws.StartTime, &ws.EndTime, &ws.BreakStart, &ws.BreakEnd, &ws.GracePeriodMinutes, &ws.OvertimePolicyID, &ws.HolidayCalendarID, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Jadwal bawaan standar jika belum ada jadwal di database
			return &WorkSchedule{
				ID:                 uuid.Nil,
				Name:               "Jadwal Standar DCISP",
				WorkingDays:        []int{1, 2, 3, 4, 5},
				StartTime:          "08:30:00",
				EndTime:            "17:00:00",
				BreakStart:         "12:00:00",
				BreakEnd:           "13:00:00",
				GracePeriodMinutes: 10,
			}, nil
		}
		return nil, fmt.Errorf("gagal mengambil jadwal kerja aktif: %w", err)
	}
	return &ws, nil
}

// Memeriksa apakah tanggal tertentu merupakan hari libur nasional atau libur operasional.
func (r *Repository) IsHoliday(ctx context.Context, targetDate time.Time) (bool, string, error) {
	query := `
		SELECT name
		FROM holidays
		WHERE date = $1::date
		LIMIT 1
	`
	var name string
	err := r.db.Pool.QueryRow(ctx, query, targetDate.Format("2006-01-02")).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("gagal memeriksa hari libur: %w", err)
	}
	return true, name, nil
}

// ============================================================================
// 3. ATTENDANCE EVENT LOGS (IMMUTABLE APPEND-ONLY)
// ============================================================================

// Menyimpan catatan log presensi fisik secara kekal dengan perlindungan idempotensi (ON CONFLICT DO NOTHING).
func (r *Repository) CreateEventLog(ctx context.Context, ev *AttendanceEventLog) (bool, error) {
	query := `
		INSERT INTO attendance_event_logs (id, user_id, device_id, event_type, timestamp, method, location_context, session_id, idempotency_key, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP)
		ON CONFLICT (idempotency_key) DO NOTHING
	`
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	if len(ev.Metadata) == 0 {
		ev.Metadata = []byte("{}")
	}

	cmdTag, err := r.db.Pool.Exec(ctx, query, ev.ID, ev.UserID, ev.DeviceID, ev.EventType, ev.Timestamp, ev.Method, ev.LocationContext, ev.SessionID, ev.IdempotencyKey, ev.Metadata)
	if err != nil {
		return false, fmt.Errorf("gagal menyimpan log presensi: %w", err)
	}

	return cmdTag.RowsAffected() > 0, nil
}

// Memeriksa apakah pengguna telah memiliki rekaman CHECK_IN yang sah pada tanggal tertentu.
func (r *Repository) HasCheckInToday(ctx context.Context, userID uuid.UUID, targetDate time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM attendance_event_logs
		WHERE user_id = $1
		  AND event_type = 'CHECK_IN'
		  AND (timestamp AT TIME ZONE 'Asia/Jakarta')::date = $2::date
	`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, userID, targetDate.Format("2006-01-02")).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("gagal memeriksa check-in harian: %w", err)
	}
	return count > 0, nil
}

// Mengambil event presensi terakhir pengguna pada hari tertentu untuk mengevaluasi mode cerdas (AUTO).
func (r *Repository) GetLatestEventForUserToday(ctx context.Context, userID uuid.UUID, targetDate time.Time) (*AttendanceEventLog, error) {
	query := `
		SELECT id, user_id, device_id, event_type, timestamp, method, location_context, session_id, idempotency_key, metadata, created_at
		FROM attendance_event_logs
		WHERE user_id = $1
		  AND (timestamp AT TIME ZONE 'Asia/Jakarta')::date = $2::date
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var ev AttendanceEventLog
	err := r.db.Pool.QueryRow(ctx, query, userID, targetDate.Format("2006-01-02")).Scan(
		&ev.ID, &ev.UserID, &ev.DeviceID, &ev.EventType, &ev.Timestamp, &ev.Method, &ev.LocationContext, &ev.SessionID, &ev.IdempotencyKey, &ev.Metadata, &ev.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mengambil event terakhir pengguna: %w", err)
	}
	return &ev, nil
}

// Mencari identitas pengguna berdasarkan nomor induk (NIM/NISN) atau kecocokan ID akun pengguna.
func (r *Repository) FindUserByIdentifier(ctx context.Context, identifier string) (uuid.UUID, string, string, error) {
	// 1. Cek di tabel interns berdasarkan id_number
	queryIntern := `
		SELECT u.id, u.full_name, u.status
		FROM interns i
		JOIN users u ON u.id = i.user_id
		WHERE i.id_number = $1
		LIMIT 1
	`
	var uID uuid.UUID
	var name, status string
	err := r.db.Pool.QueryRow(ctx, queryIntern, identifier).Scan(&uID, &name, &status)
	if err == nil {
		return uID, name, status, nil
	}

	// 2. Cek langsung di tabel users jika identifier berupa format UUID
	if parsedUUID, errParse := uuid.Parse(identifier); errParse == nil {
		queryUser := `SELECT id, full_name, status FROM users WHERE id = $1`
		err = r.db.Pool.QueryRow(ctx, queryUser, parsedUUID).Scan(&uID, &name, &status)
		if err == nil {
			return uID, name, status, nil
		}
	}

	return uuid.Nil, "", "", errors.New("kartu atau pengguna tidak terdaftar")
}

// ============================================================================
// 4. WORK SESSIONS & BREAKS
// ============================================================================

// Menyimpan sesi kerja harian baru ke dalam tabel work_sessions.
func (r *Repository) CreateWorkSession(ctx context.Context, ws *WorkSession) error {
	query := `
		INSERT INTO work_sessions (id, user_id, date, start_time, end_time, gross_duration_seconds, active_duration_seconds, idle_duration_seconds, active_task_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if ws.ID == uuid.Nil {
		ws.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, ws.ID, ws.UserID, ws.Date, ws.StartTime, ws.EndTime, ws.GrossDurationSeconds, ws.ActiveDurationSeconds, ws.IdleDurationSeconds, ws.ActiveTaskID, ws.Status)
	if err != nil {
		return fmt.Errorf("gagal membuat sesi kerja baru: %w", err)
	}
	return nil
}

// Mengambil sesi kerja aktif pengguna pada hari ini yang belum ditutup (status bukan ENDED).
func (r *Repository) GetActiveWorkSession(ctx context.Context, userID uuid.UUID, targetDate time.Time) (*WorkSession, error) {
	query := `
		SELECT id, user_id, date, start_time, end_time, gross_duration_seconds, active_duration_seconds, idle_duration_seconds, active_task_id, status, created_at, updated_at
		FROM work_sessions
		WHERE user_id = $1
		  AND date = $2::date
		  AND status != 'ENDED'
		ORDER BY created_at DESC
		LIMIT 1
	`
	var ws WorkSession
	err := r.db.Pool.QueryRow(ctx, query, userID, targetDate.Format("2006-01-02")).Scan(
		&ws.ID, &ws.UserID, &ws.Date, &ws.StartTime, &ws.EndTime, &ws.GrossDurationSeconds, &ws.ActiveDurationSeconds, &ws.IdleDurationSeconds, &ws.ActiveTaskID, &ws.Status, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mengambil sesi kerja aktif: %w", err)
	}
	return &ws, nil
}

// Memperbarui status dan durasi sesi kerja.
func (r *Repository) UpdateWorkSessionStatus(ctx context.Context, sessionID uuid.UUID, status string) error {
	query := `
		UPDATE work_sessions
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, sessionID, status)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status sesi kerja: %w", err)
	}
	return nil
}

// Menutup sesi kerja harian dengan menghitung durasi kotor, aktif, dan idle.
func (r *Repository) EndWorkSession(ctx context.Context, sessionID uuid.UUID, endTime time.Time, gross, active, idle int) error {
	query := `
		UPDATE work_sessions
		SET end_time = $2,
		    gross_duration_seconds = $3,
		    active_duration_seconds = $4,
		    idle_duration_seconds = $5,
		    status = 'ENDED',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, sessionID, endTime, gross, active, idle)
	if err != nil {
		return fmt.Errorf("gagal menutup sesi kerja: %w", err)
	}
	return nil
}

// Mencatat dimulainya periode istirahat (Break).
func (r *Repository) CreateBreak(ctx context.Context, b *Break) error {
	query := `
		INSERT INTO breaks (id, session_id, start_time, end_time, duration_seconds, is_anomaly_early, is_violation_late, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
	`
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, b.ID, b.SessionID, b.StartTime, b.EndTime, b.DurationSeconds, b.IsAnomalyEarly, b.IsViolationLate)
	if err != nil {
		return fmt.Errorf("gagal mencatat periode istirahat: %w", err)
	}
	return nil
}

// Mengambil rekaman istirahat aktif yang belum selesai untuk sesi kerja tertentu.
func (r *Repository) GetActiveBreak(ctx context.Context, sessionID uuid.UUID) (*Break, error) {
	query := `
		SELECT id, session_id, start_time, end_time, duration_seconds, is_anomaly_early, is_violation_late, created_at
		FROM breaks
		WHERE session_id = $1 AND end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`
	var b Break
	err := r.db.Pool.QueryRow(ctx, query, sessionID).Scan(
		&b.ID, &b.SessionID, &b.StartTime, &b.EndTime, &b.DurationSeconds, &b.IsAnomalyEarly, &b.IsViolationLate, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari istirahat aktif: %w", err)
	}
	return &b, nil
}

// Menutup periode istirahat dan mencatat status keterlambatan kembali bekerja.
func (r *Repository) EndBreak(ctx context.Context, breakID uuid.UUID, endTime time.Time, durationSeconds int, isLate bool) error {
	query := `
		UPDATE breaks
		SET end_time = $2, duration_seconds = $3, is_violation_late = $4
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, breakID, endTime, durationSeconds, isLate)
	if err != nil {
		return fmt.Errorf("gagal mengakhiri istirahat: %w", err)
	}
	return nil
}

// Mengambil seluruh total durasi istirahat yang telah diambil dalam sebuah sesi kerja.
func (r *Repository) GetTotalBreakDurationForSession(ctx context.Context, sessionID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(SUM(duration_seconds), 0) FROM breaks WHERE session_id = $1`
	var total int
	err := r.db.Pool.QueryRow(ctx, query, sessionID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("gagal menghitung durasi istirahat sesi: %w", err)
	}
	return total, nil
}

// ============================================================================
// 5. OVERTIME REQUESTS (FR-012)
// ============================================================================

// Menyimpan permohonan lembur baru ke dalam database.
func (r *Repository) CreateOvertime(ctx context.Context, ot *OvertimeRequest) error {
	query := `
		INSERT INTO overtime_requests (id, user_id, project_id, task_id, date, requested_start, requested_end, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if ot.ID == uuid.Nil {
		ot.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, ot.ID, ot.UserID, ot.ProjectID, ot.TaskID, ot.Date, ot.RequestedStart, ot.RequestedEnd, ot.Reason, ot.Status)
	if err != nil {
		return fmt.Errorf("gagal menyimpan permohonan lembur: %w", err)
	}
	return nil
}

// Mengambil permohonan lembur berdasarkan ID.
func (r *Repository) GetOvertimeByID(ctx context.Context, id uuid.UUID) (*OvertimeRequest, error) {
	query := `
		SELECT o.id, o.user_id, o.project_id, o.task_id, o.date, o.requested_start::text, o.requested_end::text, o.approved_start::text, o.approved_end::text, o.actual_worked_minutes, o.reason, o.status, o.reviewer_id, o.review_notes, o.created_at, o.updated_at,
		       u.full_name
		FROM overtime_requests o
		JOIN users u ON u.id = o.user_id
		WHERE o.id = $1
	`
	var ot OvertimeRequest
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&ot.ID, &ot.UserID, &ot.ProjectID, &ot.TaskID, &ot.Date, &ot.RequestedStart, &ot.RequestedEnd, &ot.ApprovedStart, &ot.ApprovedEnd, &ot.ActualWorkedMinutes, &ot.Reason, &ot.Status, &ot.ReviewerID, &ot.ReviewNotes, &ot.CreatedAt, &ot.UpdatedAt,
		&ot.UserFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari lembur berdasarkan id: %w", err)
	}
	return &ot, nil
}

// Memperbarui status persetujuan permohonan lembur oleh supervisor.
func (r *Repository) ReviewOvertime(ctx context.Context, id uuid.UUID, status string, appStart, appEnd *string, reviewerID uuid.UUID, notes *string) error {
	query := `
		UPDATE overtime_requests
		SET status = $2, approved_start = $3::time, approved_end = $4::time, reviewer_id = $5, review_notes = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id, status, appStart, appEnd, reviewerID, notes)
	if err != nil {
		return fmt.Errorf("gagal memperbarui persetujuan lembur: %w", err)
	}
	return nil
}

// Mengambil permohonan lembur yang disetujui untuk pengguna pada tanggal tertentu.
func (r *Repository) GetApprovedOvertimeToday(ctx context.Context, userID uuid.UUID, targetDate time.Time) (*OvertimeRequest, error) {
	query := `
		SELECT id, user_id, project_id, task_id, date, requested_start::text, requested_end::text, approved_start::text, approved_end::text, actual_worked_minutes, reason, status, reviewer_id, review_notes, created_at, updated_at, ''
		FROM overtime_requests
		WHERE user_id = $1 AND date = $2::date AND status = 'APPROVED'
		LIMIT 1
	`
	var ot OvertimeRequest
	err := r.db.Pool.QueryRow(ctx, query, userID, targetDate.Format("2006-01-02")).Scan(
		&ot.ID, &ot.UserID, &ot.ProjectID, &ot.TaskID, &ot.Date, &ot.RequestedStart, &ot.RequestedEnd, &ot.ApprovedStart, &ot.ApprovedEnd, &ot.ActualWorkedMinutes, &ot.Reason, &ot.Status, &ot.ReviewerID, &ot.ReviewNotes, &ot.CreatedAt, &ot.UpdatedAt, &ot.UserFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal memeriksa lembur disetujui: %w", err)
	}
	return &ot, nil
}

// ============================================================================
// 6. LEAVE REQUESTS (FR-014)
// ============================================================================

// Menyimpan permohonan cuti baru ke dalam tabel leave_requests.
func (r *Repository) CreateLeave(ctx context.Context, req *LeaveRequest) error {
	query := `
		INSERT INTO leave_requests (id, user_id, leave_type, start_date, end_date, reason, evidence_file_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, req.ID, req.UserID, req.LeaveType, req.StartDate, req.EndDate, req.Reason, req.EvidenceFileID, req.Status)
	if err != nil {
		return fmt.Errorf("gagal membuat permohonan cuti: %w", err)
	}
	return nil
}

// Mengambil data permohonan cuti berdasarkan ID.
func (r *Repository) GetLeaveByID(ctx context.Context, id uuid.UUID) (*LeaveRequest, error) {
	query := `
		SELECT l.id, l.user_id, l.leave_type, l.start_date, l.end_date, l.reason, l.evidence_file_id, l.status, l.reviewer_id, l.created_at, l.updated_at,
		       u.full_name
		FROM leave_requests l
		JOIN users u ON u.id = l.user_id
		WHERE l.id = $1
	`
	var lr LeaveRequest
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&lr.ID, &lr.UserID, &lr.LeaveType, &lr.StartDate, &lr.EndDate, &lr.Reason, &lr.EvidenceFileID, &lr.Status, &lr.ReviewerID, &lr.CreatedAt, &lr.UpdatedAt,
		&lr.UserFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari cuti berdasarkan id: %w", err)
	}
	return &lr, nil
}

// Memperbarui status persetujuan permohonan cuti oleh peninjau berwenang.
func (r *Repository) ReviewLeave(ctx context.Context, id uuid.UUID, status string, reviewerID uuid.UUID) error {
	query := `
		UPDATE leave_requests
		SET status = $2, reviewer_id = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id, status, reviewerID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status cuti: %w", err)
	}
	return nil
}

// Memeriksa apakah pengguna memiliki cuti yang disetujui pada tanggal tertentu untuk pembebasan kewajiban presensi.
func (r *Repository) HasApprovedLeaveOnDate(ctx context.Context, userID uuid.UUID, targetDate time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM leave_requests
		WHERE user_id = $1
		  AND status = 'APPROVED'
		  AND $2::date BETWEEN start_date AND end_date
	`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, userID, targetDate.Format("2006-01-02")).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("gagal memeriksa status cuti aktif: %w", err)
	}
	return count > 0, nil
}

// ============================================================================
// 7. ATTENDANCE CORRECTIONS (APPEND-ONLY - FR-015, BR-025)
// ============================================================================

// Menyimpan permohonan koreksi absensi manual secara append-only.
func (r *Repository) CreateCorrection(ctx context.Context, corr *AttendanceCorrection) error {
	query := `
		INSERT INTO attendance_corrections (id, user_id, attendance_event_id, target_date, proposed_event_type, proposed_timestamp, reason, evidence_file_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if corr.ID == uuid.Nil {
		corr.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, corr.ID, corr.UserID, corr.AttendanceEventID, corr.TargetDate, corr.ProposedEventType, corr.ProposedTimestamp, corr.Reason, corr.EvidenceFileID, corr.Status)
	if err != nil {
		return fmt.Errorf("gagal membuat koreksi absensi: %w", err)
	}
	return nil
}

// Mengambil rekaman koreksi absensi berdasarkan ID.
func (r *Repository) GetCorrectionByID(ctx context.Context, id uuid.UUID) (*AttendanceCorrection, error) {
	query := `
		SELECT c.id, c.user_id, c.attendance_event_id, c.target_date, c.proposed_event_type, c.proposed_timestamp, c.reason, c.evidence_file_id, c.status, c.reviewer_id, c.reviewed_at, c.created_at, c.updated_at,
		       u.full_name
		FROM attendance_corrections c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1
	`
	var ac AttendanceCorrection
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&ac.ID, &ac.UserID, &ac.AttendanceEventID, &ac.TargetDate, &ac.ProposedEventType, &ac.ProposedTimestamp, &ac.Reason, &ac.EvidenceFileID, &ac.Status, &ac.ReviewerID, &ac.ReviewedAt, &ac.CreatedAt, &ac.UpdatedAt,
		&ac.UserFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari koreksi absensi: %w", err)
	}
	return &ac, nil
}

// Memperbarui status persetujuan koreksi absensi dan tanggal review.
func (r *Repository) ReviewCorrection(ctx context.Context, id uuid.UUID, status string, reviewerID uuid.UUID) error {
	query := `
		UPDATE attendance_corrections
		SET status = $2, reviewer_id = $3, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id, status, reviewerID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status koreksi absensi: %w", err)
	}
	return nil
}
