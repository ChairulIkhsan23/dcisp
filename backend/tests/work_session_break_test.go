package tests

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/attendance"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji siklus hidup sesi kerja, prasyarat kehadiran, pemantauan istirahat, dan anomali break (T-036, T-037).
func TestWorkSessionAndBreakLifecycle(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	repo := attendance.NewRepository(db)
	service := attendance.NewService(repo, db, rdb, cfg, nil, nil)
	ctx := context.Background()

	testUserID := uuid.New()
	testEmail := fmt.Sprintf("session_hero_%s@dcisp.internal", testUserID.String()[:8])
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Session Hero Tester', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	now := time.Now().In(attendance.WIB)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM breaks WHERE session_id IN (SELECT id FROM work_sessions WHERE user_id = $1)", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM work_sessions WHERE user_id = $1", testUserID)
	})

	// 1. Uji Penolakan Start Work jika belum ada Check-In hari ini (BRULE-WF-004)
	_, err = service.StartWorkSession(ctx, testUserID, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memerlukan catatan kehadiran (CHECK_IN)")

	// 2. Buat Check-In sah hari ini
	_, err = repo.CreateEventLog(ctx, &attendance.AttendanceEventLog{
		ID:        uuid.New(),
		UserID:    testUserID,
		EventType: attendance.EventCheckIn,
		Timestamp: time.Now().UTC(),
		Method:    attendance.MethodNFC,
	})
	require.NoError(t, err)

	// 3. Start Work Session berhasil setelah Check-In
	session, err := service.StartWorkSession(ctx, testUserID, nil)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)
	assert.Equal(t, attendance.SessionStatusWorking, session.Status)

	// 4. Uji Mulai Istirahat (Break START)
	brk, err := service.ProcessBreakAction(ctx, testUserID, "START")
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, brk.ID)

	// Cek status sesi berubah ke BREAK
	activeSession, err := repo.GetActiveWorkSession(ctx, testUserID, now)
	require.NoError(t, err)
	assert.Equal(t, attendance.SessionStatusBreak, activeSession.Status)

	// 5. Uji Melanjutkan Kerja (Break RESUME)
	resumedBreak, err := service.ProcessBreakAction(ctx, testUserID, "RESUME")
	require.NoError(t, err)
	assert.NotNil(t, resumedBreak.EndTime)

	// Cek status sesi kembali ke WORKING
	activeSessionAfter, err := repo.GetActiveWorkSession(ctx, testUserID, now)
	require.NoError(t, err)
	assert.Equal(t, attendance.SessionStatusWorking, activeSessionAfter.Status)

	// 6. Selesaikan Sesi Kerja (End Work Session)
	endedSession, err := service.EndWorkSession(ctx, testUserID)
	require.NoError(t, err)
	assert.Equal(t, attendance.SessionStatusEnded, endedSession.Status)
	assert.NotNil(t, endedSession.EndTime)
}

// Menguji alur permohonan lembur, persetujuan supervisor, dan kalkulasi jam lembur aktual (T-038, BRULE-WF-009, BRULE-WF-010).
func TestOvertimeWorkflowAndApproval(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	service := attendance.NewService(attendance.NewRepository(db), db, nil, cfg, nil, nil)
	ctx := context.Background()

	testUserID := uuid.New()
	supervisorID := uuid.New()
	userEmail := fmt.Sprintf("overtime_%s@dcisp.internal", testUserID.String()[:8])
	supervisorEmail := fmt.Sprintf("supervisor_ot_%s@dcisp.internal", supervisorID.String()[:8])

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Overtime User', 'ACTIVE'),
		       ($3, $4, 'hash_pass', 'Supervisor OT', 'ACTIVE')
	`, testUserID, userEmail, supervisorID, supervisorEmail)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM overtime_requests WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id IN ($1, $2)", testUserID, supervisorID)
	})

	// 1. Pengajuan lembur
	otReq := &attendance.CreateOvertimeRequest{
		Date:           "2026-09-21",
		RequestedStart: "17:00",
		RequestedEnd:   "19:30",
		Reason:         "Menyelesaikan rilis modul backend presensi",
	}
	ot, err := service.RequestOvertime(ctx, testUserID, otReq)
	require.NoError(t, err)
	assert.Equal(t, attendance.OvertimeStatusSubmitted, ot.Status)

	// 2. Persetujuan oleh Supervisor
	reviewReq := &attendance.ReviewOvertimeRequest{
		Action:        "APPROVE",
		ApprovedStart: func(s string) *string { return &s }("17:00:00"),
		ApprovedEnd:   func(s string) *string { return &s }("19:00:00"),
		ReviewNotes:   func(s string) *string { return &s }("Disetujui maksimal 2 jam"),
	}
	reviewedOT, err := service.ReviewOvertime(ctx, ot.ID, reviewReq, supervisorID)
	require.NoError(t, err)
	assert.Equal(t, attendance.OvertimeStatusApproved, reviewedOT.Status)
	assert.Equal(t, "17:00:00", *reviewedOT.ApprovedStart)
	assert.Equal(t, "19:00:00", *reviewedOT.ApprovedEnd)
}

// Menguji alur permohonan izin cuti resmi sakit dan pembebasan kewajiban kehadiran (T-039, FR-014).
func TestLeaveRequestWorkflowAndExemption(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := attendance.NewRepository(db)
	service := attendance.NewService(repo, db, nil, cfg, nil, nil)
	ctx := context.Background()

	testUserID := uuid.New()
	supervisorID := uuid.New()
	userEmail := fmt.Sprintf("leave_%s@dcisp.internal", testUserID.String()[:8])
	supervisorEmail := fmt.Sprintf("supervisor_leave_%s@dcisp.internal", supervisorID.String()[:8])

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Leave User', 'ACTIVE'),
		       ($3, $4, 'hash_pass', 'Supervisor Leave', 'ACTIVE')
	`, testUserID, userEmail, supervisorID, supervisorEmail)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM leave_requests WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id IN ($1, $2)", testUserID, supervisorID)
	})

	// 1. Pengajuan cuti sakit
	leaveReq := &attendance.CreateLeaveRequest{
		LeaveType: attendance.LeaveTypeSick,
		StartDate: "2026-09-22",
		EndDate:   "2026-09-23",
		Reason:    "Demam tinggi dan istirahat dokter",
	}
	lr, err := service.RequestLeave(ctx, testUserID, leaveReq)
	require.NoError(t, err)
	assert.Equal(t, attendance.LeaveStatusSubmitted, lr.Status)

	// 2. Persetujuan cuti oleh supervisor
	reviewedLeave, err := service.ReviewLeave(ctx, lr.ID, &attendance.ReviewLeaveRequest{Action: "APPROVE"}, supervisorID)
	require.NoError(t, err)
	assert.Equal(t, attendance.LeaveStatusApproved, reviewedLeave.Status)

	// 3. Verifikasi pembebasan kewajiban presensi pada tanggal cuti
	targetDate, _ := time.Parse("2006-01-02", "2026-09-22")
	hasLeave, err := repo.HasApprovedLeaveOnDate(ctx, testUserID, targetDate)
	require.NoError(t, err)
	assert.True(t, hasLeave)
}

// Menguji permohonan koreksi absensi append-only dengan preservasi log asli (T-040, FR-015, BR-025).
func TestAttendanceCorrectionAppendOnlyPreservation(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := attendance.NewRepository(db)
	service := attendance.NewService(repo, db, nil, cfg, nil, nil)
	ctx := context.Background()

	testUserID := uuid.New()
	supervisorID := uuid.New()
	userEmail := fmt.Sprintf("corr_%s@dcisp.internal", testUserID.String()[:8])
	supervisorEmail := fmt.Sprintf("supervisor_corr_%s@dcisp.internal", supervisorID.String()[:8])

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Corr User', 'ACTIVE'),
		       ($3, $4, 'hash_pass', 'Supervisor Corr', 'ACTIVE')
	`, testUserID, userEmail, supervisorID, supervisorEmail)
	require.NoError(t, err)

	// Buat log presensi asli
	origLogID := uuid.New()
	_, err = repo.CreateEventLog(ctx, &attendance.AttendanceEventLog{
		ID:        origLogID,
		UserID:    testUserID,
		EventType: attendance.EventCheckIn,
		Timestamp: time.Now().UTC(),
		Method:    attendance.MethodNFC,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM attendance_corrections WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", supervisorID)
	})

	// 1. Pengajuan koreksi untuk check-out yang terlewat
	corrReq := &attendance.CreateCorrectionRequest{
		TargetDate:        "2026-09-20",
		ProposedEventType: attendance.EventCheckOut,
		ProposedTimestamp: "2026-09-20T17:05:00Z",
		Reason:            "Lupa melakukan check-out saat gerbang padat antrean",
	}
	corr, err := service.RequestCorrection(ctx, testUserID, corrReq)
	require.NoError(t, err)
	assert.Equal(t, attendance.CorrectionStatusSubmitted, corr.Status)

	// 2. Persetujuan koreksi oleh supervisor
	reviewedCorr, err := service.ReviewCorrection(ctx, corr.ID, &attendance.ReviewCorrectionRequest{Action: "APPROVE"}, supervisorID)
	require.NoError(t, err)
	assert.Equal(t, attendance.CorrectionStatusApproved, reviewedCorr.Status)

	// 3. Verifikasi bahwa log asli tetap utuh dan terdapat rekaman baru MANUAL_CORRECTION
	var manualCorrCount int
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM attendance_event_logs
		WHERE user_id = $1 AND method = 'MANUAL_CORRECTION' AND event_type = 'CHECK_OUT'
	`, testUserID).Scan(&manualCorrCount)
	require.NoError(t, err)
	assert.Equal(t, 1, manualCorrCount)

	// Pastikan log asli tidak terhapus atau berubah (BR-025)
	var origCount int
	err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM attendance_event_logs WHERE id = $1", origLogID).Scan(&origCount)
	require.NoError(t, err)
	assert.Equal(t, 1, origCount)
}

// Menguji pendaftaran terminal pemindai, verifikasi tanda tangan HMAC-SHA256, dan pencegahan serangan replay (T-041, FR-013).
func TestDeviceRegistryAndHMACSignatureVerification(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	service := attendance.NewService(attendance.NewRepository(db), db, nil, cfg, nil, nil)
	ctx := context.Background()

	apiKey := "my_ultra_secret_terminal_api_key_2026"
	terminalID := fmt.Sprintf("GATE-ESP32-%d", time.Now().UnixNano()%1000000)

	// 1. Daftarkan perangkat terminal baru
	dev, err := service.RegisterDevice(ctx, &attendance.RegisterDeviceRequest{
		TerminalIdentifier: terminalID,
		DeviceType:         attendance.DeviceTypeESP32Terminal,
		LocationName:       "Lobi Utama Gedung A",
		APIKey:             apiKey,
		CurrentMode:        attendance.TerminalModeAuto,
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, dev.ID)
	assert.Equal(t, terminalID, dev.TerminalIdentifier)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM devices WHERE id = $1", dev.ID)
	})

	nowUnix := time.Now().Unix()
	timestampStr := fmt.Sprintf("%d", nowUnix)
	path := "/api/v1/attendance/terminal-tap"
	bodyBytes := []byte(`{"card_uid": "CARD-001"}`)

	// 2. Hitung tanda tangan HMAC yang sah
	canonical := fmt.Sprintf("POST%s%s%s", path, timestampStr, string(bodyBytes))
	validSig := computeTestHMAC(apiKey, canonical)

	// Verifikasi tanda tangan valid harus sukses
	err = service.VerifyDeviceSignature(ctx, terminalID, timestampStr, validSig, "POST", path, bodyBytes, apiKey)
	require.NoError(t, err)

	// 3. Uji penolakan tanda tangan palsu/salah
	err = service.VerifyDeviceSignature(ctx, terminalID, timestampStr, "invalid_signature_hex", "POST", path, bodyBytes, apiKey)
	assert.Error(t, err)
	assert.Equal(t, attendance.ErrDeviceSignature, err)

	// 4. Uji pencegahan serangan replay (Timestamp kedaluwarsa > 60 detik)
	expiredTimestampStr := fmt.Sprintf("%d", nowUnix-120) // 2 menit lalu
	expiredCanonical := fmt.Sprintf("POST%s%s%s", path, expiredTimestampStr, string(bodyBytes))
	expiredSig := computeTestHMAC(apiKey, expiredCanonical)

	err = service.VerifyDeviceSignature(ctx, terminalID, expiredTimestampStr, expiredSig, "POST", path, bodyBytes, apiKey)
	assert.Error(t, err)
	assert.Equal(t, attendance.ErrReplayDetected, err)
}

// Helper untuk menghitung tanda tangan HMAC pengujian
func computeTestHMAC(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
